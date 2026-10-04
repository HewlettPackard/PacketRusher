// SPDX-License-Identifier: Apache-2.0
// © Copyright 2026 Valentin D'Emmanuele

// mongosh open5gs open5gs-subscriber.js
const ambr = { uplink: { value: 1, unit: 3 }, downlink: { value: 1, unit: 3 } };
db.subscribers.insertOne({
  schema_version: 1,
  imsi: "208930000000120",
  security: { k: "00112233445566778899AABBCCDDEEFF", opc: "00112233445566778899AABBCCDDEEFF", op: null, amf: "8000" },
  ambr: ambr,
  slice: [{
    sst: 1, sd: "010203", default_indicator: true,
    session: [{
      name: "internet", type: 1, ambr: ambr, pcc_rule: [],
      qos: { index: 9, arp: { priority_level: 8, pre_emption_capability: 1, pre_emption_vulnerability: 2 } },
    }],
  }],
  access_restriction_data: 32, network_access_mode: 0, subscriber_status: 0,
  operator_determined_barring: 0, subscribed_rau_tau_timer: 12,
});
