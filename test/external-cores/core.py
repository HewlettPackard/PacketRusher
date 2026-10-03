#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Own and join a real core's NF processes; never manage host services."""
import argparse
import json
import signal
import socket
import subprocess
import threading
import time
from pathlib import Path
from prepare import CORE_IP, DN_IP
from startup import free_registered_and_associated, open_registered_and_associated

def serve_echo(stop, errors, state):
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
            sock.bind((DN_IP, 9000))
            sock.settimeout(.2)
            while not stop.is_set():
                if (state/'dn-disable').exists():
                    sock.close()
                    (state/'dn-disabled').write_text('owned UDP DN peer closed\n')
                    stop.wait()
                    return
                try:
                    data, address = sock.recvfrom(4096)
                    sock.sendto(data, address)
                except TimeoutError:
                    pass
    except Exception as error:
        errors.append(error)
        stop.set()

def start(prefix, state, upf_prefix=None, stop_capture_when_ready=False):
    state = Path(state).resolve()
    profile = json.loads((state / "profile.json").read_text())
    prefix = Path(prefix).resolve()
    upf_prefix = Path(upf_prefix).resolve() if upf_prefix else prefix
    stop = threading.Event()
    processes, files, errors = [], [], []
    capture, echo = None, None
    for signum in (signal.SIGINT, signal.SIGTERM):
        signal.signal(signum, lambda *_: stop.set())
    try:
        if profile.get('tunnel_backend') == 'ebpf' and not profile.get('native'):
            subprocess.run(['ethtool','-K','eth0','tx','off','rx','off','tso','off','gso','off','gro','off'],check=True)
        sessions = profile.get('sessions', 1 if profile['core'] == 'open5gs' else 0)
        if profile["core"] == "open5gs":
            names = ["nrf", "udr", "udm", "ausf", "bsf", "pcf", "nssf", "upf", "smf", "amf"]
        else:
            names = ["nrf", "udr", "udm", "ausf", "pcf", "nssf"] + (["upf", "smf"] if sessions else []) + ["amf"]
        if sessions and profile.get('upf_implementation','open5gs') == 'open5gs':
            subprocess.run(["ip", "tuntap", "add", "name", "ogstun", "mode", "tun"], check=True)
            subprocess.run(["ip", "addr", "add", DN_IP + "/16", "dev", "ogstun"], check=True)
            subprocess.run(["ip", "link", "set", "ogstun", "up"], check=True)
        elif sessions:
            # The genuine UPF installs the UE subnet route on upfgtp. Give the
            # DN peer only its /32 so replies must traverse those PFCP rules.
            subprocess.run(['ip','link','add','dn0','type','dummy'],check=True)
            subprocess.run(['ip','addr','add',DN_IP+'/32','dev','dn0'],check=True)
            subprocess.run(['ip','link','set','dn0','up'],check=True)
        if sessions:
            echo = threading.Thread(target=serve_echo, args=(stop,errors,state))
            echo.start()
            capture_log = (state / "pfcp-capture.log").open("wb")
            files.append(capture_log)
            capture = subprocess.Popen(["tcpdump","-n","-U","--immediate-mode","-i","lo","-w",str(state/"pfcp-startup.pcap"),"udp","port","8805"],stdout=capture_log,stderr=subprocess.STDOUT)
            capture_deadline = time.monotonic()+5
            while not ((state/"pfcp-startup.pcap").exists() and (state/"pfcp-startup.pcap").stat().st_size >= 24):
                if capture.poll() is not None or time.monotonic() >= capture_deadline:
                    raise RuntimeError("PFCP startup capture failed")
                time.sleep(.05)
        for name in names:
            open_nf = profile['core'] == 'open5gs' or (name == 'upf' and profile['upf_implementation'] == 'open5gs')
            if open_nf:
                binary = (upf_prefix if name == 'upf' else prefix) / "bin" / f"open5gs-{name}d"
                configuration = state / "config" / f"{name}.yaml"
            else:
                binary = (upf_prefix if name == 'upf' else prefix) / "bin" / name
                configuration = state / "config" / f"{name}cfg.yaml"
            log = (state / "core" / f"{name}-stdout.log").open("wb")
            files.append(log)
            processes.append(subprocess.Popen([str(binary), "-c", str(configuration)], stdout=log, stderr=subprocess.STDOUT, cwd=state))
        deadline = time.monotonic() + 60
        remaining = set(profile["nf_addresses"])
        evidence = None
        while remaining or not evidence:
            if capture and capture.poll() is not None:
                raise RuntimeError('PFCP capture exited during readiness; no packet proof can be accepted')
            if errors:
                raise RuntimeError(f"DN peer failed: {errors[0]}")
            for name, process in zip(names, processes):
                if process.poll() is not None:
                    raise RuntimeError(f"real {name} exited during readiness: {process.returncode}")
            for name in tuple(remaining):
                try:
                    with socket.create_connection((profile["nf_addresses"][name], 8000 if profile["core"] == "free5gc" else 7777), timeout=.2):
                        remaining.remove(name)
                except OSError:
                    pass
            if not remaining:
                try:
                    evidence = open_registered_and_associated(state) if profile["core"] == "open5gs" else free_registered_and_associated(state,profile)
                except OSError:
                    pass
            if time.monotonic() >= deadline or stop.is_set():
                raise RuntimeError(f"core readiness deadline: listeners={sorted(remaining)}, registration/PFCP={evidence}")
            time.sleep(.1)
        if stop_capture_when_ready and capture:
            # Manual benchmarks retain startup PFCP proof without capturing
            # their timed data. Ordinary acceptance/CI keeps capture enabled.
            capture.send_signal(signal.SIGINT)
            if capture.wait(timeout=5) != 0:
                raise RuntimeError('PFCP startup capture failed before benchmark')
            capture = None
        (state / "core-ready").write_text(json.dumps({"nf_listeners":profile["nf_addresses"],"registration_and_pfcp":evidence},indent=2)+"\n")
        while not stop.wait(.2):
            for name, process in zip(names, processes):
                if process.poll() is not None:
                    raise RuntimeError(f"real {name} exited: {process.returncode}")
            if capture and capture.poll() is not None:
                raise RuntimeError("PFCP capture stopped unexpectedly")
        if errors:
            raise RuntimeError(f"DN peer failed: {errors[0]}")
    finally:
        stop.set()
        if echo:
            echo.join(timeout=1)
        for process in reversed(processes):
            if process.poll() is None:
                process.terminate()
        deadline = time.monotonic() + 10
        for process in processes:
            try:
                process.wait(timeout=max(.1, deadline-time.monotonic()))
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
        if capture and capture.poll() is None:
            capture.send_signal(signal.SIGINT)
            try: capture.wait(timeout=5)
            except subprocess.TimeoutExpired:
                capture.kill(); capture.wait()
        for log in files:
            log.close()

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--prefix", required=True)
    parser.add_argument("--state", required=True)
    parser.add_argument("--upf-prefix")
    parser.add_argument("--stop-capture-when-ready", action="store_true", help="manual benchmarks only; retain startup proof then stop the owned PFCP capture")
    args = parser.parse_args()
    start(args.prefix, args.state,args.upf_prefix,args.stop_capture_when_ready)
