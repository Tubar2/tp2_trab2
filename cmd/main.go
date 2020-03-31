package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	defer elapsed("Process main")()

	//Setting path to text file
	var path = "resource/"
	if len(os.Args) > 1 {
		path += os.Args[1]
	} else {
		path += "dracula.txt"
	}

	//Extracting stop words
	stops := extractStops()

	//Setting up waitGroup
	var wg sync.WaitGroup

	//Making channels for data transfer
	readChan := make(chan string)
	processedChan := make(chan string)

	//Creating goroutines and adding them to waitgroup
	wg.Add(3)
	go read(readChan, path, &wg)
	go removeStops(readChan, processedChan, stops, &wg)
	go createMap(processedChan, &wg)

	//Waiting for each go routine to finish
	wg.Wait()
}

//Function that compares extracted word to stop words
func removeStops(rChan chan string, pChan chan string, stops []string, wg *sync.WaitGroup) {
	defer elapsed("Remove Stops")()
	var isStop bool
	defer close(pChan)
	defer wg.Done()

	for {
		word, ok := <- rChan
		if ok == false {
			break
		}
		isStop = false
		for _, stop := range stops {
			word = strings.TrimSuffix(word, "\n")
			stop = strings.TrimSuffix(stop, "\n")
			if strings.ToLower(word) == stop {
				isStop = true
				break
			}
		}
		if !isStop {
			pChan <- word
		}
	}
}

//Function that creates a map
func createMap(pChan chan string, wg *sync.WaitGroup) {
	defer elapsed("Create map")()
	m:= make(map[string]int)
	defer wg.Done()

	for {
		word, ok :=<- pChan
		if ok == false {
			break
		}
		//fmt.Print("r")
		m[word]++
	}

	print25(m)
}

//Function that prints top 25 greatest counting words
func print25(m map[string]int) {
	type kv struct {
		key string
		value int
	}
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].value < ss[j].value
	})

	sss := ss[len(ss)-25:]

	for _, kv := range sss {
		fmt.Println(kv.key, kv.value)
	}
}

//Function that opens stops.txt and returns array of words
func extractStops() (stops []string) {
	defer elapsed("Extract Stops")()
	file, err := os.Open("resource/stops.txt")
	defer file.Close()
	if err != nil {
		log.Fatal("Error opening:", err)
	}
	reader := bufio.NewReader(file)

	var word string
	for {
		word, err = reader.ReadString('\n')

		stops = append(stops, word)

		//Upon reaching end, err is io.EOF
		if err != nil {
			break
		}
	}
	if err != io.EOF {
		log.Fatal("Error during reading:", err)
	}
	return
}

//Function that times program execution
func elapsed(context string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("It took %v to %s\n", time.Since(start), context)
	}
}

//Function that reads file word by word and sends it to channel
func read(rChan chan string, path string, wg *sync.WaitGroup)  {
	defer elapsed("Read main")()
	file, err := os.Open(path)
	defer file.Close()
	defer close(rChan)
	defer wg.Done()

	if err != nil {
		log.Fatal("Error opening:", err)
	}
	reader := bufio.NewReader(file)

	var line string
	for {
		line, err = reader.ReadString('\n')

		re, err2 := regexp.Compile("[^a-zA-Z0-9'\\s]")
		if err2 != nil {
			log.Fatal(err)
		}

		line = re.ReplaceAllString(line, " ")
		words := strings.Fields(line)

		for _, word := range words{
			//fmt.Print("s")
			//fmt.Println("sending:", word)
			rChan <- word
		}

		//Upon reaching end, err is io.EOF
		if err != nil {
			break
		}
	}
	if err != io.EOF {
		log.Fatal("Error during reading:", err)
	}

}
