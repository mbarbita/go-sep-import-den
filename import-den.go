package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// time conversions
// Args format
const FMTARG = "02-01-2006 15:04:05"

// Intervals format
const FMTINTERVALS = "02-01-2006 15:04"
const FMTINTERVALS2 = "02-01-2006 15:04:05"

// Time parse
const FMTPARSE = "02/01/2006 15:04:05.00"

// Time Format for MySQL TIMEDATE(2)
const FMTOUT = "2006-01-02 15:04:05.00"

// Flags
type CmdArgs struct {
	dataFileIn    string
	intervalsFile string
	test          string
	showinfo      string
}

// struct for map[string]TimeRanges
type TimeRange struct {
	start, end time.Time
}

type FileOut struct {
	count      int
	fName      string
	file       *os.File
	writer     *bufio.Writer
	start, end time.Time
}

type FCount struct {
	a, b, c int
}

var cmdArgs = new(CmdArgs)
var timeRangesMap = make(map[string]TimeRange)
var filesOutMap = make(map[string]FileOut)
var filesOutCountMap = make(map[string]FCount)
var sortedKeys []string
var testing = false
var showinfo = false

// Print line and character positions for easy slice ref
func printLineNr(st string) {
	fmt.Println("Line:", st)
	fmt.Println("Line Len:", len(st))
	for i := 0; i < len(st); i++ {
		fmt.Printf("%2v|", st[i:i+1])
	}
	fmt.Print("\n")
	for i := 0; i < len(st); i++ {
		fmt.Printf("%2v|", i)
	}
	fmt.Print("\n\n")
}

func init() {
	flag.StringVar(&cmdArgs.dataFileIn, "filein", "src.txt", "Filename in")
	flag.StringVar(&cmdArgs.intervalsFile, "intervals", "intervals.txt", "Intervals filename")
	flag.StringVar(&cmdArgs.test, "test", "false", "Testing")
	flag.StringVar(&cmdArgs.showinfo, "showinfo", "false", "show more info")

	flag.Parse()
	if cmdArgs.test == "true" {
		testing = true
	}
	if cmdArgs.showinfo == "true" {
		showinfo = true
	}
}

func main() {
	tstart := time.Now()
	fmt.Println("Started...")
	fmt.Println("intervals file:", cmdArgs.intervalsFile)
	fmt.Println("input file:", cmdArgs.dataFileIn)
	fmt.Println("testing:", cmdArgs.test)
	fmt.Println("show more info:", cmdArgs.showinfo)
	fmt.Println()

	// Open intervals file
	intervalsFile, err := os.Open(cmdArgs.intervalsFile)
	if err != nil {
		log.Println("intervals file error:")
		log.Fatal(err)
	}
	defer intervalsFile.Close()

	scanner := bufio.NewScanner(intervalsFile)
	for scanner.Scan() {

		// skip comment lines
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}
		fields := strings.Split(scanner.Text(), ",")
		if len(fields) == 1 {
			continue
		}
		if showinfo {
			fmt.Println("fields 0", fields[0])
			fmt.Println("fields 1", fields[1])
			fmt.Println("fields 2", fields[2])
		}

		var timeRange TimeRange
		timeRange.start, err = time.Parse(FMTINTERVALS2, fields[1])
		if err != nil {
			fmt.Println("parse time start intervals error:", err.Error()) // proper error handling instead of panic in your app
		}
		if showinfo {
			fmt.Println("start, end:", timeRange.start, timeRange.end)
		}
		timeRange.end, err = time.Parse(FMTINTERVALS2, fields[2])
		if err != nil {
			fmt.Println("parse time end intervals error:", err.Error()) // proper error handling instead of panic in your app
		}
		timeRangesMap[fields[0]] = timeRange

	} //intervals

	// To store the sortedKeys in slice in sorted order
	for k := range timeRangesMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	const FMTFOutA = "02-01-2006 15.04"
	const FMTFOutB = "-15.04"
	var fileOut FileOut
	for _, k := range sortedKeys {
		if showinfo {
			fmt.Println("k:", k, "start:", timeRangesMap[k].start.Format(FMTARG), "end:", timeRangesMap[k].end.Format(FMTARG))
		}

		// init counters
		filesOutCountMap[k] = FCount{0, 0, 0}
		// fmt.Println("filesOutMap:\n", filesOutCountMap[k])

		fileOut.fName = "(" + k + ") " + timeRangesMap[k].start.Format(FMTFOutA) + timeRangesMap[k].end.Format(FMTFOutB) + ".txt"
		fileOut.file, err = os.Create(fileOut.fName)
		if err != nil {
			log.Println("create file out error:")
			log.Fatal(err)
		}
		fileOut.writer = bufio.NewWriter(fileOut.file)
		fileOut.start = timeRangesMap[k].start
		fileOut.end = timeRangesMap[k].end

		defer fileOut.file.Close()
		filesOutMap[k] = fileOut
	} //range

	// output some test data to test-data.txt
	testFileOut, err := os.Create("test-data.txt")
	if err != nil {
		log.Println("create test-data.txt file error:")
		log.Fatal(err)
	}
	defer testFileOut.Close()
	testFileOutWriter := bufio.NewWriter(testFileOut)

	// Open souece csv file
	dataFileIn, err := os.Open(cmdArgs.dataFileIn)
	if err != nil {
		log.Println("open source file error:")
		log.Fatal(err)
	}
	defer dataFileIn.Close()

	// Read lines from file
	i := 1
	j := 1
	var dayIn int
	var dayCheck = -1
	var firstentry, lastentry string

	scanner = bufio.NewScanner(dataFileIn)

	//Loop over dataFileIn
	for scanner.Scan() {

		// Print a few lines for easy slices
		if i <= 3 {
			if !testing {
				if showinfo {
					fmt.Println("index:", i)
					printLineNr(scanner.Text())
				}
			}
		}

		// Test
		if testing {
			fmt.Println("index:", i)
			printLineNr(scanner.Text())
			// 10 loops
			if i == 11 {
				break
			}
			i++
			continue
		} // test

		// Skip first line - table header
		if i == 1 {
			fmt.Println("Skip line:", i, scanner.Text())
			fmt.Println()
			// put raw first line in test-data.txt
			fmt.Fprintln(testFileOut, scanner.Text())
			testFileOutWriter.Flush()
			i++
			continue
		}

		// Process line into variables
		timeTxt := scanner.Text()[1:23]
		freqTxt := scanner.Text()[25:34]
		powerTxt := scanner.Text()[35:]

		if testing {
			fmt.Println(" i, slices:", i, timeTxt, freqTxt, powerTxt)
		}

		// tz stuff
		// t, _ := time.ParseInLocation(longForm, datetime, loc)

		// Parse time
		timeIn, err := time.Parse(FMTPARSE, timeTxt)
		if err != nil {
			log.Println("time parse error:")
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// Prepare/convert variables
		// freq = strings.TrimSpace(freq)
		freq, err := strconv.ParseFloat(strings.TrimSpace(freqTxt), 4)
		if err != nil {
			log.Println("freq parse error:")
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// power = strings.TrimSpace(powerTxt)
		power, err := strconv.ParseFloat(strings.TrimSpace(powerTxt), 3)
		if err != nil {
			log.Println("power parse error:")
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// print day processed
		dayIn = timeIn.Day()
		if dayIn != dayCheck {
			fmt.Println("processing:", timeIn.Format("02-01-2006"))
			dayCheck = dayIn
		}

		// get first entry in source file
		if i == 2 {
			firstentry = scanner.Text()
		}

		// put some data in test-data.txt
		if i <= 20 {
			fmt.Fprintln(testFileOut, scanner.Text())
			fmt.Fprintln(testFileOut, fmt.Sprintf("%v,%6.4f,%5.3f", timeIn.Format(FMTOUT), freq, power))
			testFileOutWriter.Flush()
		}

		// Insert data into file

		for k, v := range filesOutMap {
			if timeIn.After(v.start.Add(-150*time.Millisecond)) && timeIn.Before(v.end.Add(150*time.Millisecond)) {
				if timeIn.Before(v.start.Add(-150*time.Millisecond)) || timeIn.Equal(v.start) {
					// fmt.Println("first")
					fmt.Fprintln(v.file, fmt.Sprintf("%v,%6.4f,%5.3f", v.start.Format(FMTOUT), freq, power))
					rc := filesOutCountMap[k]
					rc.a++
					filesOutCountMap[k] = rc
				}
				if timeIn.After(v.start) && timeIn.Before(v.end) {
					fmt.Fprintln(v.file, fmt.Sprintf("%v,%6.4f,%5.3f", timeIn.Format(FMTOUT), freq, power))
					// filesOutCountMap[k]++
					rc := filesOutCountMap[k]
					rc.b++
					filesOutCountMap[k] = rc
				}
				if timeIn.After(v.end.Add(150*time.Millisecond)) || timeIn.Equal(v.end) {
					// fmt.Println("last")
					fmt.Fprintln(v.file, fmt.Sprintf("%v,%6.4f,%5.3f", v.end.Format(FMTOUT), freq, power))
					// filesOutCountMap[k]++
					rc := filesOutCountMap[k]
					rc.c++
					filesOutCountMap[k] = rc
				}

				if showinfo {
					// TODO: fix
					fmt.Println("found for interval:", k, " index totalcount count:", i, j, filesOutCountMap[k])
					j++
				}
				v.writer.Flush()
			}
		}

		//get lastentry in sorce file
		lastentry = scanner.Text()
		i++
	} // Loop over dataFileIn end
	if err := scanner.Err(); err != nil {
		log.Println("input file scanner error:")
		log.Fatal(err)
	}

	// Print summary
	fmt.Println()
	for _, k := range sortedKeys {
		s := filesOutCountMap[k].a + filesOutCountMap[k].b + filesOutCountMap[k].c
		fmt.Printf("id: %v file name: %v lines: %v total: %v\n", k, filesOutMap[k].fName, filesOutCountMap[k], s)
	}
	fmt.Println()
	fmt.Println("total lines:", i)
	fmt.Println("first entry:", firstentry)
	fmt.Println("last entry :", lastentry)

	// fmt.Println()
	// for k, v := range filesOutMap {
	// 	fmt.Printf("k: %v v.fName: %v count: %v\n", k, v.fName, filesOutCountMap[k])
	// }
	fmt.Println()
	fmt.Println("START:", tstart.Format("15:04:05"))
	fmt.Println("END:  ", time.Now().Format("15:04:05"))
	fmt.Println("DURATION:", time.Since(tstart))
	fmt.Println("\a")

}
