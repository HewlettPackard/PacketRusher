# SPDX-License-Identifier: Apache-2.0
"""Actual production cohort cancellation must reap its separate owned groups."""
import json
import os
import signal
import subprocess
import tempfile
import unittest
from pathlib import Path


class CohortCancellation(unittest.TestCase):
    def test_actual_measure_wait_cancellation_reaps_separate_iperf_group(self):
        here=Path(__file__).resolve().parent
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);child=root/'iperf'
            child.write_text('#!/usr/bin/env python3\nimport os,time,signal\nfrom pathlib import Path\nPath(__file__).with_suffix(".pid").write_text(str(os.getpid()));signal.signal(signal.SIGTERM,signal.SIG_IGN);time.sleep(.1);os.kill(os.getppid(),signal.SIGTERM)\nwhile True:time.sleep(1)\n')
            child.chmod(0o755)
            driver=root/'driver.py'
            driver.write_text(f'''import sys,signal
sys.path.insert(0,{str(here)!r})
import run,measure
signal.signal(signal.SIGTERM,run.cancel)
try:measure.measure([{str(child)!r}],{str(root/'flow')!r},{{}},5)
except run.Cancelled:sys.exit(0)
sys.exit(1)
''')
            process=subprocess.Popen(['python3',str(driver)],env=dict(os.environ,PYTHONDONTWRITEBYTECODE='1'),start_new_session=True)
            try:
                self.assertEqual(process.wait(timeout=6),0)
                pid=int(child.with_suffix('.pid').read_text())
                self.assertFalse(Path(f'/proc/{pid}').exists())
                value=json.loads((root/'flow.measurement.json').read_text())
                self.assertFalse(value['success']);self.assertIn('signal 15',value['error'])
            finally:
                if process.poll() is None:os.killpg(process.pid,signal.SIGKILL);process.wait()
                if child.with_suffix('.pid').exists():
                    try:os.killpg(int(child.with_suffix('.pid').read_text()),signal.SIGKILL)
                    except ProcessLookupError:pass

    def test_sigterm_inside_control_reaps_capture_and_packet_rusher_groups(self):
        here=Path(__file__).resolve().parent
        fixture=here.parent if (here.parent/'native.py').is_file() else here.parent/'01-external-cores/code/test/external-cores'
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory); state=root/'cohort';state.mkdir();(root/'config').mkdir()
            (root/'profile.json').write_text('{"core":"free5gc"}')
            (root/'config/packetrusher.yaml').write_text('{"ue":{}}')
            capture=root/'tcpdump'
            capture.write_text('#!/usr/bin/env python3\nimport sys,os,time,signal\nfrom pathlib import Path\np=Path(sys.argv[sys.argv.index("-w")+1]);p.write_bytes(bytes(24));(p.parent/"capture.pid").write_text(str(os.getpid()));signal.signal(signal.SIGTERM,signal.SIG_IGN)\nwhile True:time.sleep(1)\n')
            capture.chmod(0o755)
            cli=root/'pr'
            cli.write_text('#!/usr/bin/env python3\nimport sys,os,time,signal\nfrom pathlib import Path\nif "control" in sys.argv:\n os.kill(os.getppid(),signal.SIGTERM);sys.exit(0)\np=Path(sys.argv[sys.argv.index("--control-socket")+1]);p.touch();(p.parent/"pr.pid").write_text(str(os.getpid()));signal.signal(signal.SIGTERM,signal.SIG_IGN)\nwhile True:time.sleep(1)\n')
            cli.chmod(0o755)
            driver=root/'driver.py'
            driver.write_text(f'''import sys,json,argparse
sys.path.insert(0,{str(here)!r})
import run
run.imports({str(fixture)!r})
run.probe.await_core_count=lambda *args,**kwargs:None
run.Path.is_socket=lambda path:path.exists()
old=run.native.stop_owned
run.native.stop_owned=lambda process,timeout=12:old(process,.05)
args=argparse.Namespace(state={str(state)!r},packetrusher={str(cli)!r},backend="userspace",repetition=0,processes="{{}}",rates=[50000000],seed=1,seconds=1)
sys.exit(run.cohort(args))
''')
            env=dict(os.environ,PATH=str(root)+os.pathsep+os.environ['PATH'],PYTHONDONTWRITEBYTECODE='1')
            driver_process=subprocess.Popen(['python3',str(driver)],env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,start_new_session=True)
            try:
                output,error=driver_process.communicate(timeout=8)
                self.assertEqual(driver_process.returncode,1,error.decode())
                result=json.loads((state/'result.json').read_text())
                self.assertIn('signal 15',result['error'])
                for name in ('capture','pr'):
                    pid=int((state/f'{name}.pid').read_text())
                    self.assertFalse(Path(f'/proc/{pid}').exists(),f'{name} was not reaped')
            finally:
                if driver_process.poll() is None:
                    os.killpg(driver_process.pid,signal.SIGKILL);driver_process.wait()
                    driver_process.communicate()
                for name in ('capture','pr'):
                    path=state/f'{name}.pid'
                    if path.exists():
                        try:os.killpg(int(path.read_text()),signal.SIGKILL)
                        except ProcessLookupError:pass


if __name__=='__main__':unittest.main()
