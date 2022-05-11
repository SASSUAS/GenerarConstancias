package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	s "gsheet-to-json-csv/src/services"
	u "gsheet-to-json-csv/src/utils"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/signintech/gopdf"
)

var RUTA_CSVS = "../outputs/CSVsFiles"
var OUTPUT_FOLDER = "../outputs/PDFs/"
var TEMPLATE_PDF = "ConstanciaBase.pdf"
var PDF_NAME = "Constancia"

func UNUSED(x ...interface{}) {}

type EventosRecord struct {
	NombrePersona     string
	NombreEvento      string
	FechaEvento       string
	FechaConstancia   string
	ConsecutivoEvento string
	HorasEvento       string
}

type Config struct {
	OutputFolder       string `json:"outputFolder"`
	PDFTemplate        string `json:"pdfTemplate"`
	CSVsOutputFolder   string `json:"CSVsOutputFolder"`
	FolderCSVsDownload string `json:"FolderCSVsDownload"`
	PDFName            string `json:"PDFName"`
	DescargarCSVGoogle int16  `json:"DescargarCSVGoogle"`
	OmitirValoresCSV   int16  `json:"OmitirValoresCSV"`
	UsarGORoutines     int16  `json:"UsarGORoutines"`
	NombreEvento       string `json:"NombreEvento"`
	FechaEvento        string `json:"FechaEvento"`
	FechaConstancia    string `json:"FechaConstancia"`
	HorasEvento        string `json:"HorasEvento"`
	Campos             struct {
		NombrePersona     string `json:"NombrePersona"`
		NombreEvento      string `json:"NombreEvento"`
		FechaEvento       string `json:"FechaEvento"`
		FechaConstancia   string `json:"FechaConstancia"`
		ConsecutivoEvento string `json:"ConsecutivoEvento"`
		HorasEvento       string `json:"HorasEvento"`
	} `json:"Campos"`
	GoogleSheetURL []string `json:"GoogleSheetURL"`
}

func CSVFileToMap(filePath string) (returnMap []map[string]string) {

	// read csv file
	csvfile, err := os.Open(filePath)
	u.IsError(err, "")

	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	rawCSVdata, err := reader.ReadAll()
	u.IsError(err, "")

	header := []string{} // holds first row (header)
	for lineNum, record := range rawCSVdata {
		if lineNum == 0 {
			for i := 0; i < len(record); i++ {
				header = append(header, strings.TrimSpace(record[i]))
			}
		} else {
			line := map[string]string{}
			for i := 0; i < len(record); i++ {
				line[header[i]] = record[i]
			}
			returnMap = append(returnMap, line)
		}
	}

	return returnMap
}

func CSVFileToMapOnlyRequiredFields(jsonConfig *Config) (returnMap []map[string]string) {

	// read csv file
	csvfile, err := os.Open(jsonConfig.FolderCSVsDownload)
	u.IsError(err, "")

	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	rawCSVdata, err := reader.ReadAll()
	u.IsError(err, "")

	values := reflect.ValueOf(jsonConfig.Campos)

	jsonFields := make([]interface{}, values.NumField())

	for i := 0; i < values.NumField(); i++ {
		jsonFields[i] = values.Field(i).Interface()
	}

	header := []string{} // holds first row (header)
	headerLines := []int{}
	for lineNum, record := range rawCSVdata {
		if lineNum == 0 {
			for i := 0; i < len(record); i++ {
				if len(header) == len(jsonFields) {
					break
				}
				for j := 0; j < len(jsonFields); j++ {
					if strings.TrimSpace(record[i]) == jsonFields[j] {
						header = append(header, strings.TrimSpace(record[i]))
						headerLines = append(headerLines, i)
					}
				}
			}
		} else {
			line := map[string]string{}
			for i, j := 0, 0; i < len(record); i++ {
				if j == len(headerLines) {
					break
				}
				if i == headerLines[j] {
					line[header[j]] = record[i]
					j += 1
				}
			}
			returnMap = append(returnMap, line)
		}
	}

	return returnMap
}

func CSVDataToEventStruct(datosCSVFile *[]map[string]string, jsonConfig *Config) []EventosRecord {
	var eventosList []EventosRecord
	for _, line := range *datosCSVFile {
		var rec EventosRecord
		for key, value := range line {
			if key == jsonConfig.Campos.NombrePersona {
				rec.NombrePersona = value
			} else if key == jsonConfig.Campos.NombreEvento {
				rec.NombreEvento = value
			} else if key == jsonConfig.Campos.FechaEvento {
				rec.FechaEvento = value
			} else if key == jsonConfig.Campos.FechaConstancia {
				rec.FechaConstancia = value
			} else if key == jsonConfig.Campos.ConsecutivoEvento {
				rec.ConsecutivoEvento = value
			} else if key == jsonConfig.Campos.HorasEvento {
				rec.HorasEvento = value
			}
		}
		eventosList = append(eventosList, rec)
	}
	return eventosList
}

func handleDownloadGoogleSheet(outputPath string, googleSheetURL string, timeout int64) {
	csvFilePath := outputPath
	errorResponse := s.Download(googleSheetURL,
		csvFilePath,
		timeout,
	)
	if errorResponse.Err != nil {
		u.ErrorLogger.Println(errorResponse.Message, errorResponse.Err)
		os.Exit(1)
	}
}

func printLogs(mensaje string, valor interface{}) {
	u.GeneralLogger.Printf(mensaje, valor)
	fmt.Printf(mensaje, valor)
}

func loadPDFData(datosEventoList *[]EventosRecord, nombrePDFs *string, jsonConfig *Config) {
	var wg sync.WaitGroup
	anchoPagina := 791.63
	altoPagina := 612.01
	descripcionTamano := 16
	tamanoLista := len(*datosEventoList)
	formatoSprintf := "%0" + u.GetDigitSize(tamanoLista) + "d"
	wg.Add(tamanoLista)
	for i := range *datosEventoList {

		go createConcurrentPDF(nombrePDFs, anchoPagina, altoPagina, descripcionTamano, &((*datosEventoList)[i]), jsonConfig, &formatoSprintf, i, &wg)

	}
	wg.Wait()
}
func createConcurrentPDF(nombrePDFs *string, anchoPagina float64, altoPagina float64, descripcionTamano int, rowData *EventosRecord, jsonConfig *Config, formatoSprintf *string, consecutivo int, wgrp *sync.WaitGroup) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: anchoPagina, H: altoPagina}})

	pdf.AddPage()
	u.HandleAddFonts(&pdf, "robotoRegular", "../Fonts/Roboto/Roboto-Regular.ttf")
	u.HandleAddFonts(&pdf, "robotoBold", "../Fonts/Roboto/Roboto-Bold.ttf")
	u.HandleAddFonts(&pdf, "robotoCondensed", "../Fonts/Roboto/RobotoCondensed-Regular.ttf")
	u.HandleAddFonts(&pdf, "bookmanOldStyle", "../Fonts/BOOKOSBI.ttf")

	u.HandleSetFonts(&pdf, "bookmanOldStyle", "", 25)
	tpl1 := pdf.ImportPage(TEMPLATE_PDF, 1, "/MediaBox")
	pdf.UseImportedTemplate(tpl1, 0, 0, anchoPagina, altoPagina)

	u.HandleSetText(&pdf, 180, 320, "a: "+rowData.NombrePersona)

	u.HandleSetFonts(&pdf, "robotoCondensed", "", descripcionTamano)
	u.HandleSetText(&pdf, 110, 370, "Por su participación en el taller \"")

	u.HandleSetFonts(&pdf, "robotoBold", "", descripcionTamano)
	var nombreEvento *string
	var horasEvento *string
	var fechaEvento *string

	nombreEvento = u.ValidateDataFields(&jsonConfig.NombreEvento, &rowData.NombreEvento)

	if (len(*nombreEvento) + 35) > 95 {
		indexEvento := u.GetSpaceIndex(*nombreEvento, 60)
		pdf.Text((*nombreEvento)[0:indexEvento])
		u.HandleSetText(&pdf, 110, 390, (*nombreEvento)[(indexEvento+1):len(*nombreEvento)])
	} else {
		pdf.Text(*nombreEvento)
	}

	u.HandleSetFonts(&pdf, "robotoCondensed", "", descripcionTamano)

	fechaEvento = u.ValidateDataFields(&jsonConfig.FechaEvento, &rowData.FechaEvento)

	if *fechaEvento != "" {
		pdf.Text("\", " + *fechaEvento)
	}

	horasEvento = u.ValidateDataFields(&jsonConfig.HorasEvento, &rowData.HorasEvento)

	if *horasEvento != "" {
		pdf.Text("con una duración de " + *horasEvento)
	}

	u.HandleSetText(&pdf, 250, 450, "Culiacán, Sinaloa, "+jsonConfig.FechaConstancia)

	if jsonConfig.OmitirValoresCSV == 1 {
		pdf.WritePdf(OUTPUT_FOLDER + *nombrePDFs + fmt.Sprintf(*formatoSprintf, consecutivo) + ".pdf")
	} else {
		pdf.WritePdf(OUTPUT_FOLDER + *nombrePDFs + rowData.ConsecutivoEvento + ".pdf")
	}
	wgrp.Done()
}

func createPDF(datosEventoList *[]EventosRecord, nombrePDFs *string, jsonConfig *Config) {
	anchoPagina := 791.63
	altoPagina := 612.01
	descripcionTamano := 16
	tamanoLista := len(*datosEventoList)
	formatoSprintf := "%0" + u.GetDigitSize(tamanoLista) + "d"
	pdfNumbers := 0
	for i, rowData := range *datosEventoList {
		pdf := gopdf.GoPdf{}
		pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: anchoPagina, H: altoPagina}})

		pdf.AddPage()
		u.HandleAddFonts(&pdf, "robotoRegular", "../Fonts/Roboto/Roboto-Regular.ttf")
		u.HandleAddFonts(&pdf, "robotoBold", "../Fonts/Roboto/Roboto-Bold.ttf")
		u.HandleAddFonts(&pdf, "robotoCondensed", "../Fonts/Roboto/RobotoCondensed-Regular.ttf")
		u.HandleAddFonts(&pdf, "bookmanOldStyle", "../Fonts/BOOKOSBI.ttf")

		u.HandleSetFonts(&pdf, "bookmanOldStyle", "", 25)
		tpl1 := pdf.ImportPage(TEMPLATE_PDF, 1, "/MediaBox")
		pdf.UseImportedTemplate(tpl1, 0, 0, anchoPagina, altoPagina)

		u.HandleSetText(&pdf, 180, 320, "a: "+rowData.NombrePersona)

		u.HandleSetFonts(&pdf, "robotoCondensed", "", descripcionTamano)
		u.HandleSetText(&pdf, 110, 370, "Por su participación en el taller \"")

		u.HandleSetFonts(&pdf, "robotoBold", "", descripcionTamano)
		var nombreEvento *string
		var horasEvento *string
		var fechaEvento *string

		nombreEvento = u.ValidateDataFields(&jsonConfig.NombreEvento, &rowData.NombreEvento)

		if (len(*nombreEvento) + 35) > 95 {
			indexEvento := u.GetSpaceIndex(*nombreEvento, 60)
			pdf.Text((*nombreEvento)[0:indexEvento])
			u.HandleSetText(&pdf, 110, 390, (*nombreEvento)[(indexEvento+1):len(*nombreEvento)])
		} else {
			pdf.Text(*nombreEvento)
		}

		u.HandleSetFonts(&pdf, "robotoCondensed", "", descripcionTamano)

		fechaEvento = u.ValidateDataFields(&jsonConfig.FechaEvento, &rowData.FechaEvento)

		if *fechaEvento != "" {
			pdf.Text("\", " + *fechaEvento)
		}

		horasEvento = u.ValidateDataFields(&jsonConfig.HorasEvento, &rowData.HorasEvento)

		if *horasEvento != "" {
			pdf.Text("con una duración de " + *horasEvento)
		}

		u.HandleSetText(&pdf, 250, 450, "Culiacán, Sinaloa, "+jsonConfig.FechaConstancia)

		if jsonConfig.OmitirValoresCSV == 1 {
			pdfNumbers += 1
			pdf.WritePdf(OUTPUT_FOLDER + *nombrePDFs + fmt.Sprintf(formatoSprintf, i) + ".pdf")
		} else {
			pdf.WritePdf(OUTPUT_FOLDER + *nombrePDFs + rowData.ConsecutivoEvento + ".pdf")
		}
		printLogs("La cantidad de PDFs generados fueron de: %d \n", pdfNumbers)

		if pdfNumbers == 5 {
			break
		}
	}
}

func readJsonFile(strConfig *Config, name string) {
	jsonFile, err := os.Open(name)
	u.IsError(err, "")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &strConfig)
}

func initConfiguration() Config {
	var strConfig Config
	readJsonFile(&strConfig, "config.json")
	loadConfiguration(&strConfig)
	return strConfig
}

func loadConfiguration(strConfig *Config) {
	if strConfig.CSVsOutputFolder != "" {
		RUTA_CSVS = strConfig.CSVsOutputFolder
	}
	if strConfig.OutputFolder != "" {
		OUTPUT_FOLDER = strConfig.OutputFolder
	}
	if strConfig.PDFTemplate != "" {
		TEMPLATE_PDF = strConfig.PDFTemplate
	}
	if strConfig.PDFName != "" {
		PDF_NAME = strConfig.PDFName
	}

	u.GeneralLogger.Println(strConfig)
}

func main() {
	jsonConfig := initConfiguration()

	if jsonConfig.DescargarCSVGoogle == 1 {
		u.GeneralLogger.Println("Starting Extracting Language Files from GoogleSheet - downloading csv approach..")
		handleDownloadGoogleSheet(jsonConfig.FolderCSVsDownload, jsonConfig.GoogleSheetURL[0], 5000)
	} else {
		u.GeneralLogger.Println("No se descargo nada")
	}
	var datosEventoList []EventosRecord
	mapaEvento := CSVFileToMapOnlyRequiredFields(&jsonConfig)
	datosEventoList = CSVDataToEventStruct(&mapaEvento, &jsonConfig)
	printLogs("Inicia generacion de PDFs %s \n", "")
	if jsonConfig.UsarGORoutines == 1 {
		loadPDFData(&datosEventoList, &PDF_NAME, &jsonConfig)
	} else {
		createPDF(&datosEventoList, &PDF_NAME, &jsonConfig)
	}
	printLogs("Termina generacion de PDFs %s \n", "")
}
