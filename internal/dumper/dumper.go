package dumper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

func dump(filePath string, store structs.Storage) error {
	encodedMetrics, err := serializer.EncodeMetrics(store)
	if err != nil {
		return fmt.Errorf("failed to convert metrics to json: %s", err.Error())
	}
	log.Printf("INFO dump file content:\n %s\n", encodedMetrics)
	err = ioutil.WriteFile(filePath, encodedMetrics, 0644)
	if err != nil {
		return fmt.Errorf("failed to dump metric to file %s: %s", filePath, err.Error())
	}
	return nil
}

func restore(filePath string, store structs.Storage) error {
	jsonFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open dump file %s: %s", filePath, err.Error())
	}
	metrics := []structs.Metric{}
	err = json.Unmarshal(jsonFile, &metrics)
	if err != nil {
		return fmt.Errorf("failed to unmarshall json to serializer.Metrics: %s", err.Error())
	}
	for _, m := range metrics {
		if m.MType == "gauge" {
			store.SetGauge(m.ID, *m.Value)
		} else if m.MType == "counter" {
			store.SetCounter(m.ID, *m.Delta)
		} else {
			log.Printf("WARN Failed to restore %+v: unknown metric type", m)
		}
	}
	return nil
}

func dumpData(storeFile string, store structs.Storage) {
	log.Println("INFO dump dumping data to disk")
	err := dump(storeFile, store)
	if err != nil {
		log.Printf("ERROR dump failed to save data: %s\n", err.Error())
	} else {
		log.Printf("INFO dump successfully saved data (%s)", storeFile)
	}

}

func RestoreData(storeFile string, store structs.Storage) {
	log.Println("INFO dump restore data from disk")
	err := restore(storeFile, store)
	if err != nil {
		log.Printf("ERROR dump failed to restore data: %s\n", err.Error())
	} else {
		log.Printf("INFO dump successfully restored data (%s)", storeFile)
	}
}

func Start(ctx context.Context, wg *sync.WaitGroup,
	storeInterval time.Duration, storeFile string, store structs.Storage) {
	log.Println("INFO dump starting")
	defer wg.Done()
	ticker := time.NewTicker(storeInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO dump received ctx.Done()'. Dumping and exiting")
			dumpData(storeFile, store)
			return
		case <-ticker.C:
			dumpData(storeFile, store)
		}
	}
}
