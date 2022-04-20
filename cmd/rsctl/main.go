package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/owenrumney/go-sarif/sarif"
	"github.com/reposaur/reposaur/pkg/detector"
	"github.com/reposaur/reposaur/pkg/output"
	"github.com/reposaur/reposaur/pkg/sdk"
)

func main() {
	reposaur, err := sdk.New(context.Background(), []string{"../policy"})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := fiber.New()

	app.Use(cors.New())

	app.Get("/github/*", func(c *fiber.Ctx) error {
		ghPath := c.Params("*")

		ghReq, err := http.NewRequest(http.MethodGet, "/"+ghPath, nil)
		if err != nil {
			fmt.Println(err)
			c.SendString(err.Error())
			return c.SendStatus(http.StatusInternalServerError)
		}

		ghReq.Header.Set("User-Agent", "reposaur")

		ghResp, err := reposaur.HTTPClient().Do(ghReq)
		if err != nil {
			fmt.Println(err)
			c.SendString(err.Error())
			return c.SendStatus(http.StatusInternalServerError)
		}
		defer ghResp.Body.Close()

		if ghResp.StatusCode != http.StatusOK {
			c.SendStream(ghResp.Body)
			return c.SendStatus(ghResp.StatusCode)
		}

		var respData interface{}

		dec := json.NewDecoder(ghResp.Body)
		if err := dec.Decode(&respData); err != nil {
			fmt.Println(err)
			c.SendString(err.Error())
			return c.SendStatus(http.StatusInternalServerError)
		}

		var data []interface{}

		switch i := respData.(type) {
		case map[string]interface{}:
			data = append(data, i)

		case []interface{}:
			for _, d := range i {
				data = append(data, d)
			}
		}

		var (
			wg       = sync.WaitGroup{}
			reportCh = make(chan output.Report, len(data))
		)

		wg.Add(len(data))

		for _, d := range data {
			namespace, err := detector.DetectNamespace(d)
			if err != nil {
				return err
			}

			props, err := detector.DetectReportProperties(namespace, d)
			if err != nil {
				return err
			}

			go func(namespace string, props output.ReportProperties, data interface{}) {
				r, err := reposaur.Check(c.Context(), namespace, data)
				if err != nil {
					panic(err)
				}

				r.Properties = props
				reportCh <- r

				wg.Done()
			}(namespace, props, d)
		}

		wg.Wait()
		close(reportCh)

		var reports []*sarif.Report

		for r := range reportCh {
			sarif, err := output.NewSarifReport(r)
			if err != nil {
				return err
			}

			reports = append(reports, sarif)
		}

		var resp []byte
		if len(reports) == 1 {
			resp, err = json.Marshal(reports[0])
			if err != nil {
				return err
			}
		} else {
			resp, err = json.Marshal(reports)
			if err != nil {
				return err
			}
		}

		return c.Send(resp)
	})

	app.Listen(":8080")
}
