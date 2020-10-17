package ngapConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/aper"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/lib/openapi/models"
	"strings"
)

func TraceDataToModels(traceActivation ngapType.TraceActivation) (traceData models.TraceData) {
	// TODO: finish this function when need
	return
}

func TraceDataToNgap(traceData models.TraceData, trsr string) (traceActivation ngapType.TraceActivation) {

	if len(trsr) != 4 {
		// logger.NgapLog.Warningln("Trace Recording Session Reference should be 2 octets")
		return
	}

	//NG-RAN Trace ID (left most 6 octet Trace Reference + last 2 octet Trace Recoding Session Reference)
	subStringSlice := strings.Split(traceData.TraceRef, "-")

	if len(subStringSlice) != 2 {
		// logger.NgapLog.Warningln("TraceRef format is not correct")
		return
	}

	plmnID := models.PlmnId{}
	plmnID.Mcc = subStringSlice[0][:3]
	plmnID.Mnc = subStringSlice[0][3:]
	traceID, _ := hex.DecodeString(subStringSlice[1])

	tmp := PlmnIdToNgap(plmnID)
	traceReference := append(tmp.Value, traceID...)
	trsrNgap, _ := hex.DecodeString(trsr)

	nGRANTraceID := append(traceReference, trsrNgap...)

	traceActivation.NGRANTraceID.Value = nGRANTraceID

	// Interfaces To Trace
	interfacesToTrace, _ := hex.DecodeString(traceData.InterfaceList)
	traceActivation.InterfacesToTrace.Value = aper.BitString{
		Bytes:     interfacesToTrace,
		BitLength: 8,
	}

	// Trace Collection Entity IP Address
	ngapIP := IPAddressToNgap(traceData.CollectionEntityIpv4Addr, traceData.CollectionEntityIpv6Addr)
	traceActivation.TraceCollectionEntityIPAddress = ngapIP

	// Trace Depth
	switch traceData.TraceDepth {
	case models.TraceDepth_MINIMUM:
		traceActivation.TraceDepth.Value = ngapType.TraceDepthPresentMinimum
	case models.TraceDepth_MEDIUM:
		traceActivation.TraceDepth.Value = ngapType.TraceDepthPresentMedium
	case models.TraceDepth_MAXIMUM:
		traceActivation.TraceDepth.Value = ngapType.TraceDepthPresentMaximum
	case models.TraceDepth_MINIMUM_WO_VENDOR_EXTENSION:
		traceActivation.TraceDepth.Value = ngapType.TraceDepthPresentMinimumWithoutVendorSpecificExtension
	case models.TraceDepth_MEDIUM_WO_VENDOR_EXTENSION:
		traceActivation.TraceDepth.Value = ngapType.TraceDepthPresentMediumWithoutVendorSpecificExtension
	case models.TraceDepth_MAXIMUM_WO_VENDOR_EXTENSION:
		traceActivation.TraceDepth.Value = ngapType.TraceDepthPresentMaximumWithoutVendorSpecificExtension
	}

	return
}
