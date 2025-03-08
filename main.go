package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
)

type TemplateData struct {
	PowerLines   []PowerLineData
	Switches     []SwitchData
	Transformers []TransformerData
	Result       CalculationResult
	Task1Result  bool
	Task2Result  bool
	ActiveTab    string
}

type CalculationResult struct {
	// Результати 1 завдання
	WOC  float64
	TWOC float64
	KAOC float64
	KPOC float64
	WDK  float64
	WDC  float64

	// Результати 2 завдання
	MLosses float64
}

type PowerLineData struct {
	Name string
	W    float64
	TV   float64
	M    float64
	TP   float64
}

type SwitchData struct {
	Name string
	W    float64
	TV   float64
	M    float64
	TP   float64
}

type TransformerData struct {
	Name string
	W    float64
	TV   float64
	M    float64
	TP   float64
}

var powerLines = []PowerLineData{
	{"ПЛ-100 кВ", 0.007, 10.0, 0.167, 35.0},
	{"ПЛ-35 кВ", 0.02, 8.0, 0.167, 35.0},
	{"ПЛ-10 кВ", 0.02, 10.0, 0.167, 35.0},
	{"КЛ-10 кВ (траншея)", 0.03, 44.0, 1.0, 9.0},
	{"КЛ-10 кВ (кабельний канал)", 0.005, 17.5, 1.0, 9.0},
}

var switches = []SwitchData{
	{"В-110 кВ (елегазовий)", 0.01, 30.0, 0.1, 30.0},
	{"В-10 кВ (малооливний)", 0.02, 15.0, 0.33, 15.0},
	{"В-10 кВ (вакуумний)", 0.01, 15.0, 0.33, 15.0},
}

var transformers = []TransformerData{
	{"Т-110 кВ", 0.015, 100.0, 1.0, 43.0},
	{"Т-35 кВ", 0.02, 80.0, 1.0, 28.0},
	{"Т-10кВ (кабельна мережа)", 0.005, 60.0, 0.5, 10.0},
	{"Т-10кВ (повітряна мережа)", 0.05, 60.0, 0.5, 10.0},
}

func calculateTaskOne(powerLineIndex, switchIndex, transformerIndex int, powerLineLength, accession, switchCount float64) CalculationResult {
	pl := powerLines[powerLineIndex]
	sw := switches[switchIndex]
	tr := transformers[transformerIndex]

	wOC := (pl.W * powerLineLength) + sw.W + tr.W + (0.02 * switchCount) + (accession * 0.03)
	tWOC := ((pl.W*powerLineLength)*pl.TV + sw.W*sw.TV + tr.W*tr.TV + (0.02 * 15.0 * switchCount) + accession*0.03*2.0) / wOC
	kAOC := wOC * tWOC / 8760.0

	kPMax := math.Max(math.Max(pl.M*pl.TP, sw.M*sw.TP), tr.M*tr.TP)
	kPOC := 1.2 * (kPMax / 8760.0)
	wDK := 2.0 * wOC * (kAOC + kPOC)
	wDC := wDK + 0.02

	return CalculationResult{
		WOC:  wOC,
		TWOC: tWOC,
		KAOC: kAOC,
		KPOC: kPOC,
		WDK:  wDK,
		WDC:  wDC,
	}
}

func calculateTaskTwo(emergencyLosses, plannedLosses float64) CalculationResult {
	const (
		omega = 0.01
		tB    = 0.045
		pM    = 5120.0
		tM    = 6451.0
		kP    = 0.004
	)

	mWEmergency := omega * tB * pM * tM
	mWPlanned := kP * pM * tM
	mLosses := (emergencyLosses * mWEmergency) + (plannedLosses * mWPlanned)

	return CalculationResult{MLosses: mLosses}
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))

	data := TemplateData{
		PowerLines:   powerLines,
		Switches:     switches,
		Transformers: transformers,
		ActiveTab:    "task1", // Активна вкладка за замовчуванням
	}

	if r.Method == http.MethodPost {
		r.ParseForm()

		// Оновлення активної вкладки, якщо необхідно
		if tab := r.FormValue("activeTab"); tab != "" {
			data.ActiveTab = tab
		}

		if r.FormValue("task") == "1" {
			powerLineIndex, _ := strconv.Atoi(r.FormValue("powerLine"))
			switchIndex, _ := strconv.Atoi(r.FormValue("switch"))
			transformerIndex, _ := strconv.Atoi(r.FormValue("transformer"))
			powerLineLength, _ := strconv.ParseFloat(r.FormValue("powerLineLength"), 64)
			accession, _ := strconv.ParseFloat(r.FormValue("accession"), 64)
			switchCount, _ := strconv.ParseFloat(r.FormValue("switchCount"), 64)

			data.Result = calculateTaskOne(powerLineIndex, switchIndex, transformerIndex, powerLineLength, accession, switchCount)
			data.Task1Result = true
		}

		if r.FormValue("task") == "2" {
			emergencyLosses, _ := strconv.ParseFloat(r.FormValue("emergencyLosses"), 64)
			plannedLosses, _ := strconv.ParseFloat(r.FormValue("plannedLosses"), 64)

			data.Result = calculateTaskTwo(emergencyLosses, plannedLosses)
			data.Task2Result = true
		}
	}

	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
