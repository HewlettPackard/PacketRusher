// SPDX-License-Identifier: Apache-2.0
// © Copyright 2026 Valentin D'Emmanuele

// mongosh free5gc free5gc-subscriber.js
const ue = { ueId: "imsi-208930000000120", servingPlmnId: "20893" };
const slice = { sst: 1, sd: "010203" };
db["subscriptionData.authenticationData.authenticationSubscription"].insertOne({
  ueId: ue.ueId, authenticationMethod: "5G_AKA", authenticationManagementField: "8000",
  encPermanentKey: "00112233445566778899AABBCCDDEEFF", encOpcKey: "00112233445566778899AABBCCDDEEFF",
  sequenceNumber: { sqnScheme: "GENERAL", sqn: "000000000020" },
});
db["subscriptionData.provisionedData.amData"].insertOne({
  ...ue, gpsis: ["msisdn-0900000000"], nssai: { defaultSingleNssais: [slice], singleNssais: [slice] },
  subscribedUeAmbr: { uplink: "1 Gbps", downlink: "1 Gbps" },
});
db["subscriptionData.provisionedData.smfSelectionSubscriptionData"].insertOne({
  ...ue, subscribedSnssaiInfos: { "01010203": { dnnInfos: [{ dnn: "internet", defaultDnnIndicator: true }] } },
});
db["policyData.ues.amData"].insertOne({ ueId: ue.ueId, subscCats: ["free5gc"] });
