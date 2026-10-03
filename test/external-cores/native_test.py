# SPDX-License-Identifier: Apache-2.0
import os
from pathlib import Path
import signal
import subprocess
import sys
import time
import unittest
from native import launch_owned, stop_owned, defer_cancellation, interrupt


class OwnedCleanup(unittest.TestCase):
    def test_repeated_cancellation_reaps_both_cohorts_and_preserves_original_error(self):
        code='import signal,time; signal.signal(signal.SIGTERM,signal.SIG_IGN); print("ready",flush=True); time.sleep(60)'
        processes=[launch_owned([sys.executable,'-c',code],stdout=subprocess.PIPE,text=True) for _ in range(2)]
        previous=signal.getsignal(signal.SIGINT)
        try:
            for process in processes:
                self.assertEqual(process.stdout.readline().strip(),'ready')
            signal.signal(signal.SIGINT,interrupt)
            first_wait=processes[0].wait
            def second_cancellation(*args,**kwargs):
                os.kill(os.getpid(),signal.SIGINT)
                return first_wait(*args,**kwargs)
            processes[0].wait=second_cancellation
            with self.assertRaisesRegex(InterruptedError,'original cancellation'):
                try:
                    raise InterruptedError('original cancellation')
                finally:
                    with defer_cancellation():
                        for process in processes:
                            stop_owned(process,timeout=.1)
            self.assertEqual([process.returncode for process in processes],[-signal.SIGKILL,-signal.SIGKILL])
            self.assertIs(signal.getsignal(signal.SIGINT),interrupt)
        finally:
            signal.signal(signal.SIGINT,previous)
            for process in processes:
                stop_owned(process,timeout=.1)
                process.stdout.close()

    def test_timeout_retires_descendant_even_when_leader_already_exited(self):
        # The child remains in our owned session after its leader exits.
        code = 'import os,signal,time; child=os.fork(); print(child,flush=True) if child else None; exit(0) if child else signal.signal(signal.SIGTERM,signal.SIG_IGN); time.sleep(60)'
        process = launch_owned([sys.executable,'-c',code],stdout=subprocess.PIPE,text=True)
        try:
            child = int(process.stdout.readline())
            process.wait(timeout=3)
            self.assertEqual(os.getpgid(child),process.pid)
            stop_owned(process,timeout=.1)
            deadline = time.monotonic()+3
            while time.monotonic() < deadline:
                try:
                    status = Path(f'/proc/{child}/stat').read_text().split()[2]
                except FileNotFoundError:
                    break
                if status == 'Z':
                    break
                time.sleep(.01)
            else:
                self.fail('owned descendant survived bounded cleanup')
        finally:
            stop_owned(process,timeout=.1)
            process.stdout.close()

if __name__=='__main__': unittest.main()
