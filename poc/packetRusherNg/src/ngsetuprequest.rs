/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
use crate::*;

pub fn create_NGSetupRequest() -> NGAP_PDU {
    let mut ies = vec![];

    ies.push(create_globalRANNode());
    ies.push(create_RANNodeName());
    ies.push(create_supportedTAList());
    ies.push(create_defaultPagingDRX());

    NGAP_PDU::InitiatingMessage(InitiatingMessage {
        procedure_code: ProcedureCode(21),
        criticality: Criticality(Criticality::REJECT),
        value: InitiatingMessageValue::Id_NGSetup(NGSetupRequest {
            protocol_i_es: NGSetupRequestProtocolIEs(ies),
        }),
    })
}

fn create_defaultPagingDRX() -> NGSetupRequestProtocolIEs_Entry {
    let ie_entry = NGSetupRequestProtocolIEs_Entry {
        id: ProtocolIE_ID(21),
        criticality: Criticality(Criticality::REJECT),
        value: NGSetupRequestProtocolIEs_EntryValue::Id_DefaultPagingDRX(PagingDRX(PagingDRX::V128)),
    };
    ie_entry
}

fn create_supportedTAList() -> NGSetupRequestProtocolIEs_Entry {
    let ie_entry = NGSetupRequestProtocolIEs_Entry {
        id: ProtocolIE_ID(102),
        criticality: Criticality(Criticality::REJECT),
        value: NGSetupRequestProtocolIEs_EntryValue::Id_SupportedTAList(SupportedTAList(vec![SupportedTAItem {
            tac: TAC(vec![0, 0, 1]),
            broadcast_plmn_list: BroadcastPLMNList(vec![BroadcastPLMNItem {
                plmn_identity: PLMNIdentity(vec![0x99, 0xf9, 0x72]),
                tai_slice_support_list: SliceSupportList(vec![SliceSupportItem {
                    s_nssai: S_NSSAI {
                        sst: SST(vec![1]),
                        sd: Some(SD(vec![0, 0, 1])),
                        ie_extensions: None,
                    },
                    ie_extensions: None
                }]),
                ie_extensions: None,
            }]),
            ie_extensions: None,
        }])),
    };
    ie_entry
}

fn create_RANNodeName() -> NGSetupRequestProtocolIEs_Entry {
    let ie_entry = NGSetupRequestProtocolIEs_Entry {
        id: ProtocolIE_ID(82),
        criticality: Criticality(Criticality::REJECT),
        value: NGSetupRequestProtocolIEs_EntryValue::Id_RANNodeName(RANNodeName("PacketRusherNG".to_string())),
    };
    ie_entry
}

fn create_globalRANNode() -> NGSetupRequestProtocolIEs_Entry {
    let ie_entry = NGSetupRequestProtocolIEs_Entry {
        id: ProtocolIE_ID(27),
        criticality: Criticality(Criticality::REJECT),
        value: NGSetupRequestProtocolIEs_EntryValue::Id_GlobalRANNodeID(
            GlobalRANNodeID::GlobalGNB_ID(
                GlobalGNB_ID {
                    plmn_identity: PLMNIdentity(vec![0x99, 0xf9, 0x72]),
                    gnb_id: GNB_ID::GNB_ID(GNB_ID_gNB_ID(BitVec::from_bitslice(bits![u8, Msb0; 1; 24]))),
                    ie_extensions: None,
                })
        ),
    };
    ie_entry
}
