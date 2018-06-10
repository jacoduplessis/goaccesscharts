package main

import (
	"bytes"
	"encoding/json"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/svg"
	"github.com/wcharczuk/go-chart"
	"html/template"
	"log"
	"os"
	"time"
)

type CountMaxMin struct {
	Count int
	Max   int
	Min   int
}

type CountPercent struct {
	Count   float64
	Percent float64
}

type Metadata struct {
	Visitors CountMaxMin
	Hits     CountMaxMin
	Data     struct {
		Unique int
	}
}

type HitsVisitorsData []struct {
	Hits     CountPercent
	Visitors CountPercent
	Data     string
}

type Visitors struct {
	Metadata Metadata
	Data     HitsVisitorsData // date in form "20180610"
}

type Requests struct {
	Metadata Metadata
	Data     HitsVisitorsData // path that was visited e.g. "/home/path/"
}

type Report struct {
	General struct {
		StartDate     string `json:"start_date"`
		EndDate       string `json:"end_date"`
		DateTime      string `json:"date_time"`
		TotalRequests int    `json:"total_requests"`
	}
	Visitors Visitors
	Requests Requests
	Hosts    struct {
		Metadata Metadata
		Data     []struct {
			Hits     CountPercent
			Visitors CountPercent
			Data     string
			Country  string
		}
	}
	OS struct {
		Metadata Metadata
		Data     []struct {
			Hits     CountPercent
			Visitors CountPercent
			Data     string
			Items    HitsVisitorsData
		}
	}
}

type TemplateContext struct {
	Visitors template.HTML
}

func getBaseChart() *chart.Chart {
	return &chart.Chart{
		// Width: 1920, //this overrides the default.
		Background: chart.Style{
			Padding: chart.Box{
				Top:  40,
				Left: 80,
			},
		},

		Series: []chart.Series{},
		XAxis: chart.XAxis{
			Name:      "Date",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxisSecondary: chart.YAxis{
			Name:      "Hits",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Name:      "Visitors",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
	}

}

func getVisitorsChart(v Visitors) *chart.Chart {
	var y1, y2 []float64
	var x []time.Time
	for _, d := range v.Data {
		y1 = append(y1, d.Visitors.Count)
		y2 = append(y2, d.Hits.Count)
		t, err := time.Parse("20060102", d.Data)
		if err != nil {
			log.Fatal(err)
		}
		x = append(x, t)
	}

	c := getBaseChart()

	c.Series = []chart.Series{
		chart.TimeSeries{
			XValues: x,
			YValues: y1,
		},
		chart.TimeSeries{
			XValues: x,
			YValues: y2,
			YAxis:   chart.YAxisSecondary,
		},
	}
	return c
}

func ChartAsHTML(c *chart.Chart) template.HTML {
	markup := bytes.Buffer{}
	c.Render(chart.SVG, &markup)
	m := minify.New()
	m.AddFunc("image/svg", svg.Minify)
	gb, _ := m.Bytes("image/svg", markup.Bytes())
	return template.HTML(gb)

}

func main() {

	var d Report

	err := json.NewDecoder(os.Stdin).Decode(&d)
	if err != nil {
		log.Fatal(err)
	}
	t := template.Must(template.ParseFiles("template.html"))

	ctx := TemplateContext{
		Visitors: ChartAsHTML(getVisitorsChart(d.Visitors)),
	}
	err = t.ExecuteTemplate(os.Stdout, "template.html", &ctx)
	if err != nil {
		log.Fatal(err)
	}
}
