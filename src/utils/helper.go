package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/signintech/gopdf"
)

// ErrorResponse exported
type ErrorResponse struct {
	Message string
	Err     error
}

// GeneralLogger exported
var GeneralLogger *log.Logger

// ErrorLogger exported
var ErrorLogger *log.Logger

func init() {
	absPath, err := filepath.Abs("../outputs/log")
	IsError(err, "Error reading given path:")

	generalLog, err := os.OpenFile(absPath+"/general-log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if IsError(err, "Error opening file:") {
		os.Exit(1)
	}

	GeneralLogger = log.New(generalLog, "General Logger:\t", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	ErrorLogger = log.New(generalLog, "Error Logger:\t", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}
func GetDigitSize(numero int) string {
	var count int
	for numero > 0 {
		numero = numero / 10
		count++
	}
	return strconv.Itoa(count)
}

func GetSpaceIndex(palabra string, tamano int32) int {
	anterior := 0
	for i := 0; i < len(palabra); i++ {
		if string(palabra[i]) == " " {
			if i > int(tamano) {
				return anterior
			} else if i == int(tamano) {
				return i
			} else {
				anterior = i
			}
		}
	}
	return anterior
}

func deleteFile(path *string) {
	var err = os.Remove(*path)
	if IsError(err, "") {
		return
	}
}

func IsError(err error, descripcion string) bool {
	if err != nil {
		fmt.Println(descripcion + err.Error())
	}
	return (err != nil)
}

// ReturnErrorResponse exported
func ReturnErrorResponse(err error, message string) *ErrorResponse {
	return &ErrorResponse{
		Message: message,
		Err:     err,
	}
}

func HandleSetFonts(self *gopdf.GoPdf, fontName string, style string, size int) {
	var err = self.SetFont(fontName, style, size)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleSetText(self *gopdf.GoPdf, x float64, y float64, texto string) {
	self.SetX(x)
	self.SetY(y)
	self.Text(texto)
}

func ValidateDataFields(datosConfig *string, datosEvento *string) *string {
	if *datosConfig == "" {
		return datosEvento
	} else {
		return datosConfig
	}
}

func HandleAddFonts(self *gopdf.GoPdf, fontName string, route string) {
	var err = self.AddTTFFont(fontName, route)
	if err != nil {
		log.Fatal(err)
	}
}

/*func readFolderFiles() []string {
	var nombreArchivos []string

	file, err := os.Open(RUTA_CSVS)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()
	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	for _, name := range list {
		//fmt.Print(filepath.Ext(name))
		if filepath.Ext(name) == ".csv" {
			nombreArchivos = append(nombreArchivos, name)
		}
	}
	return nombreArchivos
}*/

/*func ObtenerDatosEventoRecord(CSVFolder string) []EventosRecord {
	f, err := os.Open(CSVFolder)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var eventosList []EventosRecord
	for i, line := range data {
		if i > 0 {
			var rec EventosRecord
			for j, field := range line {
				if j == 0 {
					rec.NombrePersona = field
				} else if j == 1 {
					rec.NombreEvento = field
				} else if j == 2 {
					rec.FechaEvento = field
				} else if j == 3 {
					rec.ConsecutivoEvento = field
				} else if j == 4 {
					rec.HorasEvento = field
				}
			}
			eventosList = append(eventosList, rec)
		}
	}
	return eventosList
}*/
