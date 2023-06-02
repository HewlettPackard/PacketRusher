package main

var CmdTree = CmdNodeList{
	CmdToken{
		Name: "add",
		Next: CmdNodeList{
			CmdToken{
				Name: "pdr",
				Next: CmdFunc(CmdAddPDR),
			},
			CmdToken{
				Name: "far",
				Next: CmdFunc(CmdAddFAR),
			},
			CmdToken{
				Name: "qer",
				Next: CmdFunc(CmdAddQER),
			},
			CmdToken{
				Name: "urr",
				Next: CmdFunc(CmdAddURR),
			},
		},
	},
	CmdToken{
		Name: "mod",
		Next: CmdNodeList{
			CmdToken{
				Name: "pdr",
				Next: CmdFunc(CmdModPDR),
			},
			CmdToken{
				Name: "far",
				Next: CmdFunc(CmdModFAR),
			},
			CmdToken{
				Name: "qer",
				Next: CmdFunc(CmdModQER),
			},
			CmdToken{
				Name: "urr",
				Next: CmdFunc(CmdModURR),
			},
		},
	},
	CmdToken{
		Name: "delete",
		Next: CmdNodeList{
			CmdToken{
				Name: "pdr",
				Next: CmdFunc(CmdDeletePDR),
			},
			CmdToken{
				Name: "far",
				Next: CmdFunc(CmdDeleteFAR),
			},
			CmdToken{
				Name: "qer",
				Next: CmdFunc(CmdDeleteQER),
			},
			CmdToken{
				Name: "urr",
				Next: CmdFunc(CmdDeleteURR),
			},
		},
	},
	CmdToken{
		Name: "get",
		Next: CmdNodeList{
			CmdToken{
				Name: "pdr",
				Next: CmdFunc(CmdGetPDR),
			},
			CmdToken{
				Name: "far",
				Next: CmdFunc(CmdGetFAR),
			},
			CmdToken{
				Name: "qer",
				Next: CmdFunc(CmdGetQER),
			},
			CmdToken{
				Name: "urr",
				Next: CmdFunc(CmdGetURR),
			},
		},
	},
	CmdToken{
		Name: "list",
		Next: CmdNodeList{
			CmdToken{
				Name: "pdr",
				Next: CmdFunc(CmdListPDR),
			},
			CmdToken{
				Name: "far",
				Next: CmdFunc(CmdListFAR),
			},
			CmdToken{
				Name: "qer",
				Next: CmdFunc(CmdListQER),
			},
			CmdToken{
				Name: "urr",
				Next: CmdFunc(CmdListURR),
			},
		},
	},
}
