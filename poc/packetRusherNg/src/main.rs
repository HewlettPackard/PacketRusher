/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
mod ngap;
mod ngsetuprequest;
mod gnb;

use asn1_codecs::{aper::AperCodec, PerCodecData};
use bitvec::prelude::*;
use sctp_rs::NotificationOrData;
use ngap::*;

#[tokio::main(flavor = "current_thread")]
async fn main() -> std::io::Result<()> {
    let server_address: std::net::SocketAddr = "192.168.11.10:38412".parse().unwrap();

    let client_socket = sctp_rs::Socket::new_v4(sctp_rs::SocketToAssociation::OneToOne)?;

    let (connected, assoc_id) = client_socket.sctp_connectx(&[server_address]).await?;
    eprintln!("[gNB][SCTP] Connected!");

    let mut gnb = gnb::Gnb {
        socket: connected,
        info: sctp_rs::SendInfo {
            sid: 0,
            flags: 0,
            ppid: 60u32.to_be(),
            context: 0,
            assoc_id,
        },
        stateMachine: gnb::GnbState::DISCONNECTED
    };

    let msg = ngsetuprequest::create_NGSetupRequest();
    gnb.send_ngap(msg).await;
    gnb.stateMachine = gnb::GnbState::NG_SETUP_SENT;

    while let received = gnb.recv_ngap().await {
        gnb.handle(received).await;
    }

    Ok(())
}
