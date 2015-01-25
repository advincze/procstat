package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var pid = flag.Int("pid", -1, "pid under examination")
var html = flag.String("html", "", "html file to plot the results")
var tick = flag.Duration("tick", time.Second, "the tick size to record the parameters")

func main() {
	flag.Parse()

	type record struct {
		Ts  int64
		Cpu float64
		Mem float64
		Rss int64
		Vsz int64
	}
	records := make([]record, 0, 10e3)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		if *html == "" {
			os.Exit(0)
		}
		f, err := os.Create(*html)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		err = tmpl.Execute(f, struct{ Records []record }{records[:]})
		if err != nil {
			log.Fatal(err)
		}
		exec.Command("open", *html).Run()

		os.Exit(0)
	}()

	for ts := range time.NewTicker(*tick).C {
		data, _ := exec.Command("ps", "-p", strconv.Itoa(*pid), "-o", "%cpu,%mem,rss,vsz").Output()
		if len(data) == 0 {
			break
		}
		s := string(data)

		ss := strings.Fields(s[strings.Index(s, "\n")+3 : len(s)])
		cpu, _ := strconv.ParseFloat(ss[0], 64)
		mem, _ := strconv.ParseFloat(ss[1], 64)
		rss, _ := strconv.ParseInt(ss[2], 10, 64)
		vsz, _ := strconv.ParseInt(ss[3], 10, 64)
		fmt.Printf("ts: %.3f, cpu: %v%%, mem: %v%%, rss: %v MiB, vsz: %v MiB\n", float64(ts.UnixNano())/float64(10e8), cpu, mem, rss/1024, vsz/1024)
		records = append(records, record{ts.UnixNano() / 10e5, cpu, mem, rss, vsz})
	}

}

var tmpl = template.Must(template.New("tmpl").Parse(tmplStr))

const tmplStr = ` <html>
  <head>
    <script type="text/javascript"
          src="https://www.google.com/jsapi?autoload={
            'modules':[{
              'name':'visualization',
              'version':'1',
              'packages':['corechart']
            }]
          }"></script>

    <script type="text/javascript">
      google.setOnLoadCallback(drawChart);

      function drawChart() {

      var dataTable = new google.visualization.DataTable();
      dataTable.addColumn('datetime', 'Time', 'tm');
      dataTable.addColumn('number', 'cpu %', 'sl');
      dataTable.addColumn('number', 'mem %', 'mem');
      dataTable.addColumn('number', 'rss KiB', 'rss');
      dataTable.addColumn('number', 'vsz KiB', 'vsz');

      dataTable.addRows([
      	{{ range  .Records }}[new Date({{.Ts}}),  {{.Cpu}}, {{.Mem}}, {{.Rss}}, {{.Vsz}}],
      	{{ end }}]);

        var dataView = new google.visualization.DataView(dataTable);

        var options = {
          curveType: 'function',
          legend: { position: 'bottom' },
          hAxis: { format:'HH:mm:ss'}
        };

        dataView.setColumns([0,1]);
        new google.visualization.LineChart(document.getElementById('cpu_chart')).draw(dataView, options);
        
        dataView.setColumns([0,2]);
        new google.visualization.LineChart(document.getElementById('mem_chart')).draw(dataView, options);

        dataView.setColumns([0,3]);
        new google.visualization.LineChart(document.getElementById('rss_chart')).draw(dataView, options);

        dataView.setColumns([0,4]);
        new google.visualization.LineChart(document.getElementById('vsz_chart')).draw(dataView, options);
        
      }
    </script>
  </head>
  <body>
    <div id="cpu_chart" style="width: 100%; height: 500px"></div>
    <div id="mem_chart" style="width: 100%; height: 500px"></div>
    <div id="rss_chart" style="width: 100%; height: 500px"></div>
    <div id="vsz_chart" style="width: 100%; height: 500px"></div>
  </body>
</html>`
