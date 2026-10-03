# SPDX-License-Identifier: Apache-2.0
import json
import tempfile
import unittest
from pathlib import Path
from compose import compose


class ContainerOwnership(unittest.TestCase):
    def test_ebpf_has_only_scoped_caps_and_keeps_default_seccomp(self):
        with tempfile.TemporaryDirectory() as directory:
            compose('open5gs',directory,'ebpf')
            services=json.loads((Path(directory)/'compose.json').read_text())['services']
            self.assertEqual(services['ran']['cap_add'],['NET_ADMIN','BPF'])
            self.assertEqual(services['core']['cap_add'],['NET_ADMIN'])
            for service in services.values():
                self.assertNotIn('privileged',service)
                self.assertNotIn('security_opt',service)
                self.assertNotIn('network_mode',service)

    def test_registration_only_does_not_add_tun_or_bpf_permissions(self):
        with tempfile.TemporaryDirectory() as directory:
            compose('free5gc',directory)
            service=json.loads((Path(directory)/'compose.json').read_text())['services']['ran']
            self.assertNotIn('devices',service)
            self.assertNotIn('cap_add',service)
        with tempfile.TemporaryDirectory() as directory:
            with self.assertRaises(ValueError): compose('free5gc',directory,'ebpf')


if __name__=='__main__': unittest.main()
