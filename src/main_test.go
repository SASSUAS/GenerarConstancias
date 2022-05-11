package main

import (
	"fmt"
	"testing"
)

var jsonConfig Config
var mapaEvento []map[string]string
var datosEventoList []EventosRecord

func init() {
	jsonConfig = initConfiguration()
	mapaEvento = CSVFileToMapOnlyRequiredFields(&jsonConfig)
	datosEventoList = CSVDataToEventStruct(&mapaEvento, &jsonConfig)
}

func BenchmarkCreatePDF(b *testing.B) {

	b.Run(fmt.Sprintf("BenchmarkCreatePDF"), func(b *testing.B) {
		createPDF(&datosEventoList, PDF_NAME, &jsonConfig)
	})
}

func BenchmarkLoadPDFData(a *testing.B) {

	a.Run(fmt.Sprintf("BenchmarkLoadPDFData"), func(a *testing.B) {
		loadPDFData(&datosEventoList, &PDF_NAME, &jsonConfig)
	})
}

//go.exe test -benchmem -run=^$ -bench ^BenchmarkCreatePDF$ gsheet-to-json-csv/src
