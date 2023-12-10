/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
use crate::*;

pub enum GnbState {
    DISCONNECTED,
    NG_SETUP_SENT,
    NG_SETUP_SUCCESSFUL
}

pub struct Gnb {
    pub socket: sctp_rs::ConnectedSocket,
    pub info: sctp_rs::SendInfo,
    pub stateMachine: GnbState
}

impl Gnb {
    pub async fn send_ngap(&self, data: NGAP_PDU) {
        eprintln!("[gNB][NGAP] Sending: {:#?}", &data);

        let mut encode_codec_data = PerCodecData::new_aper();
        let result = data.aper_encode(&mut encode_codec_data);

        let send_data = sctp_rs::SendData {
            payload: encode_codec_data.into_bytes(),
            snd_info: Some(self.info.clone())
        };
        self.socket.sctp_send(send_data).await.expect("TODO: panic message");

    }

    pub async fn recv_ngap(&self) -> NGAP_PDU {
        match self.socket.sctp_recv().await.unwrap() {
            NotificationOrData::Notification(notify) => panic!("Notification {:#?}", notify),
            NotificationOrData::Data(data) => {
                let mut codec_data = PerCodecData::from_slice_aper(&data.payload);
                let msg = NGAP_PDU::aper_decode(&mut codec_data).unwrap();

                eprintln!("[gNB][NGAP] Receiving: {:#?}", &msg);
                msg
            }
        }
    }

    pub async fn handle(&mut self, pdu: NGAP_PDU) {
        match self.stateMachine {
            GnbState::DISCONNECTED => {

            }
            GnbState::NG_SETUP_SENT => {
                match pdu {
                    NGAP_PDU::InitiatingMessage(_) => panic!("[gNB][NGAP] Unexpected message from AMF"),
                    NGAP_PDU::SuccessfulOutcome(_) => {
                        self.stateMachine = GnbState::NG_SETUP_SUCCESSFUL;
                    }
                    NGAP_PDU::UnsuccessfulOutcome(_) => panic!("[gNB][NGAP] gNB unable to attach to AMF")
                }
            }
            GnbState::NG_SETUP_SUCCESSFUL => {}
        }
    }
}
