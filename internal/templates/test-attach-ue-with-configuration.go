package templates

func TestAttachUeWithConfiguration(tunnelEnabled bool) {
	TestMultiUesInQueue(1, tunnelEnabled, true, false, 500, 0, 1)
}
