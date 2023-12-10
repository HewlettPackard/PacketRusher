struct AuthenticationRequest {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		AuthenticationRequestMessageIdentity: AuthenticationRequestMessageIdentity,
		SpareHalfOctetAndNgksi: SpareHalfOctetAndNgksi,
		ABBA: ABBA,
		AuthenticationParameterRAND: Option<AuthenticationParameterRAND>,
		AuthenticationParameterAUTN: Option<AuthenticationParameterAUTN>,
		EAPMessage: Option<EAPMessage>,
}

struct AuthenticationResult {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		AuthenticationResultMessageIdentity: AuthenticationResultMessageIdentity,
		SpareHalfOctetAndNgksi: SpareHalfOctetAndNgksi,
		EAPMessage: EAPMessage,
		ABBA: Option<ABBA>,
}

struct AuthenticationReject {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		AuthenticationRejectMessageIdentity: AuthenticationRejectMessageIdentity,
		EAPMessage: Option<EAPMessage>,
}

struct RegistrationAccept {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		RegistrationAcceptMessageIdentity: RegistrationAcceptMessageIdentity,
		RegistrationResult5GS: RegistrationResult5GS,
		GUTI5G: Option<GUTI5G>,
		EquivalentPlmns: Option<EquivalentPlmns>,
		TAIList: Option<TAIList>,
		AllowedNSSAI: Option<AllowedNSSAI>,
		RejectedNSSAI: Option<RejectedNSSAI>,
		ConfiguredNSSAI: Option<ConfiguredNSSAI>,
		NetworkFeatureSupport5GS: Option<NetworkFeatureSupport5GS>,
		PDUSessionStatus: Option<PDUSessionStatus>,
		PDUSessionReactivationResult: Option<PDUSessionReactivationResult>,
		PDUSessionReactivationResultErrorCause: Option<PDUSessionReactivationResultErrorCause>,
		LADNInformation: Option<LADNInformation>,
		MICOIndication: Option<MICOIndication>,
		NetworkSlicingIndication: Option<NetworkSlicingIndication>,
		ServiceAreaList: Option<ServiceAreaList>,
		T3512Value: Option<T3512Value>,
		Non3GppDeregistrationTimerValue: Option<Non3GppDeregistrationTimerValue>,
		T3502Value: Option<T3502Value>,
		EmergencyNumberList: Option<EmergencyNumberList>,
		ExtendedEmergencyNumberList: Option<ExtendedEmergencyNumberList>,
		SORTransparentContainer: Option<SORTransparentContainer>,
		EAPMessage: Option<EAPMessage>,
		NSSAIInclusionMode: Option<NSSAIInclusionMode>,
		OperatordefinedAccessCategoryDefinitions: Option<OperatordefinedAccessCategoryDefinitions>,
		NegotiatedDRXParameters: Option<NegotiatedDRXParameters>,
		Non3GppNwPolicies: Option<Non3GppNwPolicies>,
		EPSBearerContextStatus: Option<EPSBearerContextStatus>,
}

struct RegistrationReject {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		RegistrationRejectMessageIdentity: RegistrationRejectMessageIdentity,
		Cause5GMM: Cause5GMM,
		T3346Value: Option<T3346Value>,
		T3502Value: Option<T3502Value>,
		EAPMessage: Option<EAPMessage>,
}

struct DLNASTransport {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		DLNASTRANSPORTMessageIdentity: DLNASTRANSPORTMessageIdentity,
		SpareHalfOctetAndPayloadContainerType: SpareHalfOctetAndPayloadContainerType,
		PayloadContainer: PayloadContainer,
		PduSessionID2Value: Option<PduSessionID2Value>,
		AdditionalInformation: Option<AdditionalInformation>,
		Cause5GMM: Option<Cause5GMM>,
		BackoffTimerValue: Option<BackoffTimerValue>,
}

struct DeregistrationAcceptUEOriginatingDeregistration {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		DeregistrationAcceptMessageIdentity: DeregistrationAcceptMessageIdentity,
}

struct DeregistrationAcceptUETerminatedDeregistration {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		DeregistrationAcceptMessageIdentity: DeregistrationAcceptMessageIdentity,
}

struct ServiceAccept {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		ServiceAcceptMessageIdentity: ServiceAcceptMessageIdentity,
		PDUSessionStatus: Option<PDUSessionStatus>,
		PDUSessionReactivationResult: Option<PDUSessionReactivationResult>,
		PDUSessionReactivationResultErrorCause: Option<PDUSessionReactivationResultErrorCause>,
		EAPMessage: Option<EAPMessage>,
}

struct ConfigurationUpdateCommand {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		ConfigurationUpdateCommandMessageIdentity: ConfigurationUpdateCommandMessageIdentity,
		ConfigurationUpdateIndication: Option<ConfigurationUpdateIndication>,
		GUTI5G: Option<GUTI5G>,
		TAIList: Option<TAIList>,
		AllowedNSSAI: Option<AllowedNSSAI>,
		ServiceAreaList: Option<ServiceAreaList>,
		FullNameForNetwork: Option<FullNameForNetwork>,
		ShortNameForNetwork: Option<ShortNameForNetwork>,
		LocalTimeZone: Option<LocalTimeZone>,
		UniversalTimeAndLocalTimeZone: Option<UniversalTimeAndLocalTimeZone>,
		NetworkDaylightSavingTime: Option<NetworkDaylightSavingTime>,
		LADNInformation: Option<LADNInformation>,
		MICOIndication: Option<MICOIndication>,
		NetworkSlicingIndication: Option<NetworkSlicingIndication>,
		ConfiguredNSSAI: Option<ConfiguredNSSAI>,
		RejectedNSSAI: Option<RejectedNSSAI>,
		OperatordefinedAccessCategoryDefinitions: Option<OperatordefinedAccessCategoryDefinitions>,
		SMSIndication: Option<SMSIndication>,
}

struct IdentityRequest {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		IdentityRequestMessageIdentity: IdentityRequestMessageIdentity,
		SpareHalfOctetAndIdentityType: SpareHalfOctetAndIdentityType,
}

struct Notification {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		NotificationMessageIdentity: NotificationMessageIdentity,
		SpareHalfOctetAndAccessType: SpareHalfOctetAndAccessType,
}

struct SecurityModeCommand {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		SecurityModeCommandMessageIdentity: SecurityModeCommandMessageIdentity,
		SelectedNASSecurityAlgorithms: SelectedNASSecurityAlgorithms,
		SpareHalfOctetAndNgksi: SpareHalfOctetAndNgksi,
		ReplayedUESecurityCapabilities: ReplayedUESecurityCapabilities,
		IMEISVRequest: Option<IMEISVRequest>,
		SelectedEPSNASSecurityAlgorithms: Option<SelectedEPSNASSecurityAlgorithms>,
		Additional5GSecurityInformation: Option<Additional5GSecurityInformation>,
		EAPMessage: Option<EAPMessage>,
		ABBA: Option<ABBA>,
		ReplayedS1UESecurityCapabilities: Option<ReplayedS1UESecurityCapabilities>,
}

struct SecurityModeReject {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		SecurityModeRejectMessageIdentity: SecurityModeRejectMessageIdentity,
		Cause5GMM: Cause5GMM,
}

struct Status5GMM {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		SpareHalfOctetAndSecurityHeaderType: SpareHalfOctetAndSecurityHeaderType,
		STATUSMessageIdentity5GMM: STATUSMessageIdentity5GMM,
		Cause5GMM: Cause5GMM,
}

struct PDUSessionEstablishmentAccept {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		PDUSESSIONESTABLISHMENTACCEPTMessageIdentity: PDUSESSIONESTABLISHMENTACCEPTMessageIdentity,
		SelectedSSCModeAndSelectedPDUSessionType: SelectedSSCModeAndSelectedPDUSessionType,
		AuthorizedQosRules: AuthorizedQosRules,
		SessionAMBR: SessionAMBR,
		Cause5GSM: Option<Cause5GSM>,
		PDUAddress: Option<PDUAddress>,
		RQTimerValue: Option<RQTimerValue>,
		SNSSAI: Option<SNSSAI>,
		AlwaysonPDUSessionIndication: Option<AlwaysonPDUSessionIndication>,
		MappedEPSBearerContexts: Option<MappedEPSBearerContexts>,
		EAPMessage: Option<EAPMessage>,
		AuthorizedQosFlowDescriptions: Option<AuthorizedQosFlowDescriptions>,
		ExtendedProtocolConfigurationOptions: Option<ExtendedProtocolConfigurationOptions>,
		DNN: Option<DNN>,
}

struct PDUSessionAuthenticationCommand {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		PDUSESSIONAUTHENTICATIONCOMMANDMessageIdentity: PDUSESSIONAUTHENTICATIONCOMMANDMessageIdentity,
		EAPMessage: EAPMessage,
		ExtendedProtocolConfigurationOptions: Option<ExtendedProtocolConfigurationOptions>,
}

struct PDUSessionAuthenticationResult {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		PDUSESSIONAUTHENTICATIONRESULTMessageIdentity: PDUSESSIONAUTHENTICATIONRESULTMessageIdentity,
		EAPMessage: Option<EAPMessage>,
		ExtendedProtocolConfigurationOptions: Option<ExtendedProtocolConfigurationOptions>,
}

struct PDUSessionModificationReject {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		PDUSESSIONMODIFICATIONREJECTMessageIdentity: PDUSESSIONMODIFICATIONREJECTMessageIdentity,
		Cause5GSM: Cause5GSM,
		BackoffTimerValue: Option<BackoffTimerValue>,
		CongestionReattemptIndicator5GSM: Option<CongestionReattemptIndicator5GSM>,
		ExtendedProtocolConfigurationOptions: Option<ExtendedProtocolConfigurationOptions>,
}

struct PDUSessionModificationComplete {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity: PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity,
		ExtendedProtocolConfigurationOptions: Option<ExtendedProtocolConfigurationOptions>,
}

struct PDUSessionReleaseRequest {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		PDUSESSIONRELEASEREQUESTMessageIdentity: PDUSESSIONRELEASEREQUESTMessageIdentity,
		Cause5GSM: Option<Cause5GSM>,
		ExtendedProtocolConfigurationOptions: Option<ExtendedProtocolConfigurationOptions>,
}

struct PDUSessionReleaseCommand {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		PDUSESSIONRELEASECOMMANDMessageIdentity: PDUSESSIONRELEASECOMMANDMessageIdentity,
		Cause5GSM: Cause5GSM,
		BackoffTimerValue: Option<BackoffTimerValue>,
		EAPMessage: Option<EAPMessage>,
		CongestionReattemptIndicator5GSM: Option<CongestionReattemptIndicator5GSM>,
		ExtendedProtocolConfigurationOptions: Option<ExtendedProtocolConfigurationOptions>,
}

struct Status5GSM {
		ExtendedProtocolDiscriminator: ExtendedProtocolDiscriminator,
		PDUSessionID: PDUSessionID,
		PTI: PTI,
		STATUSMessageIdentity5GSM: STATUSMessageIdentity5GSM,
		Cause5GSM: Cause5GSM,
}

