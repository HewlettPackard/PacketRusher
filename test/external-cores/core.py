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
from startup import free_registered, open_registered_and_associated

def serve_echo(stop, errors):
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
            sock.bind((DN_IP, 9000))
            sock.settimeout(.2)
            while not stop.is_set():
                try:
                    data, address = sock.recvfrom(4096)
                    sock.sendto(data, address)
                except TimeoutError:
                    pass
    except Exception as error:
        errors.append(error)
        stop.set()

def start(prefix, state):
    state = Path(state).resolve()
    profile = json.loads((state / "profile.json").read_text())
    prefix = Path(prefix).resolve()
    stop = threading.Event()
    processes, files, errors = [], [], []
    capture, echo = None, None
    for signum in (signal.SIGINT, signal.SIGTERM):
        signal.signal(signum, lambda *_: stop.set())
    try:
        if profile["core"] == "open5gs":
            subprocess.run(["ip", "tuntap", "add", "name", "ogstun", "mode", "tun"], check=True)
            subprocess.run(["ip", "addr", "add", DN_IP + "/16", "dev", "ogstun"], check=True)
            subprocess.run(["ip", "link", "set", "ogstun", "up"], check=True)
            names = ["nrf", "udr", "udm", "ausf", "bsf", "pcf", "nssf", "upf", "smf", "amf"]
            echo = threading.Thread(target=serve_echo, args=(stop,errors))
            echo.start()
            capture_log = (state / "pfcp-capture.log").open("wb")
            files.append(capture_log)
            capture = subprocess.Popen(["tcpdump","-n","-U","-i","lo","-w",str(state/"pfcp-startup.pcap"),"udp","port","8805"],stdout=capture_log,stderr=subprocess.STDOUT)
            capture_deadline = time.monotonic()+5
            while not ((state/"pfcp-startup.pcap").exists() and (state/"pfcp-startup.pcap").stat().st_size >= 24):
                if capture.poll() is not None or time.monotonic() >= capture_deadline:
                    raise RuntimeError("PFCP startup capture failed")
                time.sleep(.05)
        else:
            names = ["nrf", "udr", "udm", "ausf", "pcf", "nssf", "amf"]
        for name in names:
            if profile["core"] == "open5gs":
                binary = prefix / "bin" / f"open5gs-{name}d"
                configuration = state / "config" / f"{name}.yaml"
            else:
                binary = prefix / "bin" / name
                configuration = state / "config" / f"{name}cfg.yaml"
            log = (state / "core" / f"{name}-stdout.log").open("wb")
            files.append(log)
            processes.append(subprocess.Popen([str(binary), "-c", str(configuration)], stdout=log, stderr=subprocess.STDOUT, cwd=state))
        deadline = time.monotonic() + 60
        remaining = set(profile["nf_addresses"])
        evidence = None
        while remaining or not evidence:
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
                    evidence = open_registered_and_associated(state) if profile["core"] == "open5gs" else free_registered(profile)
                except OSError:
                    pass
            if time.monotonic() >= deadline or stop.is_set():
                raise RuntimeError(f"core readiness deadline: listeners={sorted(remaining)}, registration/PFCP={evidence}")
            time.sleep(.1)
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
    args = parser.parse_args()
    start(args.prefix, args.state)
