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
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

func encodeMetrics() ([]byte, error) {
	m := serializer.Metrics{}
	for k, v := range storage.Server.GetAllCounters() {
		m = append(m, serializer.Metric{ID: k, MType: "counter", Delta: &v})
	}
	for k, v := range storage.Server.GetAllGauges() {
		m = append(m, serializer.Metric{ID: k, MType: "gauge", Value: &v})
	}

	json, err := serializer.EncodeMetrics(m)
	if err != nil {
		return []byte{}, err
	}
	return json, nil
}

func dump(filePath string) error {
	encodedMetrics, err := encodeMetrics()
	if err != nil {
		return fmt.Errorf("failed to convert metrics to json: %s", err.Error())
	}
	err = ioutil.WriteFile(filePath, encodedMetrics, 0644)
	if err != nil {
		return fmt.Errorf("failed to dump metric to file %s: %s", filePath, err.Error())
	}
	return nil
}

func restore(filePath string) error {
	jsonFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open dump file %s: %s", filePath, err.Error())
	}
	metrics := serializer.Metrics{}
	err = json.Unmarshal(jsonFile, &metrics)
	if err != nil {
		return fmt.Errorf("failed to unmarshall json to serializer.Metrics: %s", err.Error())
	}
	for _, m := range metrics {
		if m.MType == "gauge" {
			storage.Server.SetGauge(m.ID, *m.Value)
		} else if m.MType == "counter" {
			storage.Server.SetCounter(m.ID, *m.Delta)
		}
	}
	return nil
}

func dumpData(storeFile string) {
	log.Println("INFO dump dumping data to disk")
	err := dump(storeFile)
	if err != nil {
		log.Printf("ERROR dump failed to save data: %s\n", err.Error())
	} else {
		log.Printf("INFO dump successfully saved data (%s)", storeFile)
	}

}

func restoreData(storeFile string) {
	log.Println("INFO dump restore data from disk")
	err := restore(storeFile)
	if err != nil {
		log.Printf("ERROR dump failed to restore data: %s\n", err.Error())
	} else {
		log.Printf("INFO dump successfully saved data (%s)", storeFile)
	}
}

func Start(ctx context.Context, wg *sync.WaitGroup, storeInterval time.Duration, storeFile string, restore bool) {
	defer wg.Done()
	if restore {
		restoreData(storeFile)
	}
	ticker := time.NewTicker(storeInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO dumper received ctx.Done()'. Dumping and exiting")
			dumpData(storeFile)
			return
		case <-ticker.C:
			dumpData(storeFile)
		}
	}
}
