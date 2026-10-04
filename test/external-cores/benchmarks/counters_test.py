# SPDX-License-Identifier: Apache-2.0
import copy
import json
import tempfile
import unittest
from pathlib import Path
from counters import NAMES, OwnedCounters, decode_dump, delta


def dump(values=(10,20,0,3)):
    return [{'key':[f'0x{byte:02x}' for byte in index.to_bytes(4,'little')],'value':[f'0x{byte:02x}' for byte in value.to_bytes(8,'little')]} for index,value in enumerate(values)]


def stat(start=100):
    fields=['0']*20;fields[19]=str(start)
    return '42 (PacketRusher owned context) '+' '.join(fields)


class KernelCounterOwnership(unittest.TestCase):
    def test_verified_raw_byte_abi_rejects_partial_duplicate_and_wrong_width(self):
        self.assertEqual(decode_dump(dump()),dict(zip(NAMES.values(),(10,20,0,3))))
        for rows in [dump()[:-1],dump()+dump()[:1],[{'key':[0],'value':[0]*8}],{'formatted':{'key':0,'value':10}}]:
            with self.assertRaises(AssertionError):decode_dump(rows)

    def test_only_map_ids_from_exact_owner_are_queried_and_owner_loss_stops_dump(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);info=root/'42/fdinfo';info.mkdir(parents=True)
            descriptor=info/'8';descriptor.write_text('pos:\t0\nmap_id:\t17\n')
            (info.parent/'stat').write_text(stat())
            calls=[]
            def runner(command,**kwargs):
                calls.append(command)
                if command[3]=='show':
                    return json.dumps({'id':17,'name':'counters','type':'array','bytes_key':4,'bytes_value':8,'max_entries':4})
                return json.dumps(dump())
            source=OwnedCounters(42,root,runner)
            snapshot=source.snapshot()
            self.assertEqual(snapshot['map_id'],17)
            self.assertEqual(snapshot['owner_start_ticks'],100)
            self.assertTrue(all(command[-1]=='17' for command in calls))
            before=len(calls);descriptor.unlink()
            with self.assertRaisesRegex(AssertionError,'no longer owns'):source.snapshot()
            self.assertEqual(len(calls),before)

    def test_named_but_wrong_abi_or_reused_pid_cannot_supply_counters(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);info=root/'42/fdinfo';info.mkdir(parents=True)
            (info/'8').write_text('map_id:\t17\n');(info.parent/'stat').write_text(stat())
            metadata={'id':17,'name':'counters','type':'array','bytes_key':4,'bytes_value':8,'max_entries':4}
            for replacement in [{'max_entries':3},{'id':18},{'type':'percpu_array'},{'bytes_value':4}]:
                with self.assertRaises(AssertionError):OwnedCounters(42,root,lambda *a,**k:json.dumps(metadata|replacement))
            source=OwnedCounters(42,root,lambda *a,**k:json.dumps(metadata))
            (info.parent/'stat').write_text(stat(200))
            with self.assertRaises(AssertionError):source.snapshot()

    def test_deltas_never_mask_reset_or_change_in_exact_map_owner(self):
        before={'owner_pid':42,'owner_start_ticks':100,'map_id':17,'values':decode_dump(dump())}
        after=copy.deepcopy(before);after['values']=decode_dump(dump((110,220,1,6)))
        self.assertEqual(delta(before,after),dict(zip(NAMES.values(),(100,200,1,3))))
        for change in [{'map_id':18},{'owner_start_ticks':200},{'values':decode_dump(dump((0,20,0,3)))}]:
            with self.assertRaises(AssertionError):delta(before,after|change)


if __name__=='__main__':unittest.main()
