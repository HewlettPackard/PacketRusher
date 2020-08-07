package nasConvert

import (
	"my5G-RANTester/lib/openapi/models"
)

func LadnToModels(buf []uint8) (dnnValues []string) {

	for bufOffset := 1; bufOffset < len(buf); {
		lenOfDnn := int(buf[bufOffset])
		dnn := string(buf[bufOffset : bufOffset+lenOfDnn])
		dnnValues = append(dnnValues, dnn)
		bufOffset += lenOfDnn
	}

	return
}

func LadnToNas(dnn string, taiLists []models.Tai) (ladnNas []uint8) {

	dnnNas := []byte(dnn)

	ladnNas = append(ladnNas, uint8(len(dnnNas)))
	ladnNas = append(ladnNas, dnnNas...)

	taiListNas := TaiListToNas(taiLists)
	ladnNas = append(ladnNas, uint8(len(taiListNas)))
	ladnNas = append(ladnNas, taiListNas...)
	return
}
