package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Strd struct {
	XMLName xml.Name `xml:"Strd"`
	Cd      string   `xml:"RfrdDocInf>Tp>CdOrPrtry>Cd"`
	Nb      string   `xml:"RfrdDocInf>Nb"`
	RltdDt  string   `xml:"RfrdDocInf>RltdDt"`
	RmtdAmt RmtdAmt  `xml:"RmtdAmt"`
}

type RmtdAmt struct {
	RmtdAmt string `xml:",chardata"`
	RAName  string `xml:"name,attr"`
}

type Document struct {
	XMLName  xml.Name `xml:"Document"`
	Chequeno string   `xml:"CstmrCdtTrfInitn>PmtInf>CdtTrfTxInf>ChqInstr>ChqNb"`
}

func main() {
	// Open our xmlFile
	xmlFile, err := os.Open("test2.xml")
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our Document
	var doc Document
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'doc' which we defined above
	xml.Unmarshal(byteValue, &doc)

	csvfiles, err := filepath.Glob("*.csv")
	if err != nil {
		log.Fatal(err)
	}
	csvFile, _ := os.Open(csvfiles[0])
	reader := csv.NewReader(bufio.NewReader(csvFile))

	var x []Strd
	var st Strd
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		if line[0] == doc.Chequeno {
			st = Strd{
				Nb:     line[0],
				RltdDt: line[1],
				Cd:     "CINV",
			}
			st.RmtdAmt = RmtdAmt{RAName: "Ccy", RmtdAmt: line[2]}
			x = append(x, st)
		}
	}

	xmlfiles, err := filepath.Glob("*.xml")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(xmlfiles[0], os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	defer f.Close()

	output, err := xml.MarshalIndent(x, "", "\t")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	y := []byte(output)
	z := bytes.SplitAfter(y, []byte("\n"))

	var c []byte
	for i := 0; i < len(z); i++ {
		c = append(c, z[i]...)
		c = append(c, []byte("\t\t\t\t")...)
	}

	findval := []byte("</Cdtr>")
	splitstr := bytes.SplitAfter(byteValue, findval)
	newline := []byte("\n\t\t\t\t")

	byteValue = append(splitstr[0], newline...)
	byteValue = append(byteValue, c[:len(c)-4]...)
	byteValue = append(byteValue, splitstr[1]...)
	err = ioutil.WriteFile("Result.xml", byteValue, 0644)

}
