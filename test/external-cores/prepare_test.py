# SPDX-License-Identifier: Apache-2.0
import json
import tempfile
import unittest
from pathlib import Path
from prepare import generate, profile_options, UE_IP


class RealProfiles(unittest.TestCase):
    def test_registration_only_cannot_claim_backend_or_upf_coverage(self):
        self.assertEqual(profile_options('free5gc'),(0,'userspace',None))
        for options in [dict(backend='ebpf'),dict(upf='free5gc')]:
            with self.assertRaises(ValueError): profile_options('free5gc',**options)

    def test_kernel_backend_is_explicit_native_only_and_still_requires_a_pdu(self):
        with self.assertRaises(ValueError): profile_options('free5gc',sessions=1,backend='gtp5g')
        with self.assertRaises(ValueError): profile_options('free5gc',backend='gtp5g',native=True)
        self.assertEqual(profile_options('free5gc',sessions=1,backend='gtp5g',native=True),(1,'gtp5g','free5gc'))

    def test_genuine_free5gc_has_real_smf_upf_and_static_subscribed_ue(self):
        with tempfile.TemporaryDirectory() as directory:
            generate('free5gc',directory,native=True,sessions=1,backend='ebpf')
            root=Path(directory)
            profile=json.loads((root/'profile.json').read_text())
            smf=json.loads((root/'config/smfcfg.yaml').read_text())['configuration']
            upf=json.loads((root/'config/upfcfg.yaml').read_text())
            self.assertEqual(profile['upf_implementation'],'free5gc')
            self.assertEqual(profile['pfcp_smf_ip'],'127.0.0.18')
            self.assertNotIn('upf',profile['nf_addresses'])
            self.assertEqual(smf['plmnList'],[{'mcc':'208','mnc':'93'}])
            self.assertEqual(smf['snssaiInfos'][0]['sNssai']['sd'],'010203')
            self.assertEqual(upf['gtpu']['forwarder'],'gtp5g')
            self.assertEqual(json.loads((root/'config/uerouting.yaml').read_text())['ueRoutingInfo'],{})
            self.assertFalse(smf['nwInstFqdnEncoding'])
            js=(root/'subscriber.js').read_text()
            self.assertIn('"staticIpAddress": [{"ipv4Addr": "'+UE_IP+'"}]',js)
            self.assertIn('policyData.ues.smData',js)

    def test_hybrid_is_explicit_and_uses_real_open5gs_upf_configuration(self):
        with tempfile.TemporaryDirectory() as directory:
            generate('free5gc',directory,native=True,sessions=1,backend='ebpf',upf='open5gs')
            root=Path(directory)
            profile=json.loads((root/'profile.json').read_text())
            smf=json.loads((root/'config/smfcfg.yaml').read_text())['configuration']
            self.assertEqual(profile['upf_implementation'],'open5gs')
            self.assertTrue(smf['nwInstFqdnEncoding'])
            self.assertTrue((root/'config/upf.yaml').is_file())
            self.assertFalse((root/'config/upfcfg.yaml').exists())


if __name__=='__main__': unittest.main()
