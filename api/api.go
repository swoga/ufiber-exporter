package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/swoga/ufiber-exporter/config"
	"github.com/swoga/ufiber-exporter/model"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 5,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Duration(5 * time.Minute),
	}
)

func request(ctx context.Context, log zerolog.Logger, device config.Device, auth string, method string, url string, data interface{}) (res *http.Response, err error) {
	var buf io.Reader
	if data != nil {
		body, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(body)
	}

	url = fmt.Sprintf("https://%s/api/v1.0/%s", device.Address, url)
	log.Debug().Str("method", method).Str("url", url).Msg("send request")

	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return
	}

	if auth != "" {
		req.Header.Add("X-Auth-Token", auth)
	}

	res, err = httpClient.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = errors.New("non-200 response")
	}

	if err != nil {
		data, _ := ioutil.ReadAll(res.Body)
		log.Error().Str("response", string(data)).Msg("error from API")
	}

	return
}

func DoLogin(ctx context.Context, log zerolog.Logger, device config.Device) (res *http.Response, err error) {
	login := &model.LoginRequest{
		Username: *device.Username,
		Password: *device.Password,
	}

	res, err = request(ctx, log, device, "", "POST", "user/login", login)
	if err != nil {
		return
	}

	var data model.LoginResponse

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return
	}

	log.Debug().Interface("data", data).Msg("response")

	defer res.Body.Close()

	if data.Error != 0 {
		err = errors.New("error != 0")
	}

	return
}

func GetStatistics(ctx context.Context, log zerolog.Logger, device config.Device, auth string) (*model.Statistics, error) {
	res, err := request(ctx, log, device, auth, "GET", "statistics", nil)
	if err != nil {
		return nil, err
	}

	var data []model.Statistics

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	log.Debug().Interface("data", data).Msg("response")

	defer res.Body.Close()

	return &data[0], nil
}

func GetInterfaces(ctx context.Context, log zerolog.Logger, device config.Device, auth string) (*[]model.InterfacesInterface, error) {
	res, err := request(ctx, log, device, auth, "GET", "interfaces", nil)
	if err != nil {
		return nil, err
	}

	var data []model.InterfacesInterface

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	log.Debug().Interface("data", data).Msg("response")

	defer res.Body.Close()

	return &data, nil
}

func GetONUs(ctx context.Context, log zerolog.Logger, device config.Device, auth string) (*[]model.ONU, error) {
	res, err := request(ctx, log, device, auth, "GET", "gpon/onus", nil)
	if err != nil {
		return nil, err
	}

	var data []model.ONU

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	log.Debug().Interface("data", data).Msg("response")

	defer res.Body.Close()

	return &data, nil
}

func GetONUsSettings(ctx context.Context, log zerolog.Logger, device config.Device, auth string) (*[]model.ONUSettings, error) {
	res, err := request(ctx, log, device, auth, "GET", "gpon/onus/settings", nil)
	if err != nil {
		return nil, err
	}

	var data []model.ONUSettings

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	log.Debug().Interface("data", data).Msg("response")

	defer res.Body.Close()

	return &data, nil
}

func GetMACTable(ctx context.Context, log zerolog.Logger, device config.Device, auth string) (*[]model.MACTable, error) {
	res, err := request(ctx, log, device, auth, "GET", "tools/mac-table", nil)
	if err != nil {
		return nil, err
	}

	var data []model.MACTable

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	log.Debug().Interface("data", data).Msg("response")

	defer res.Body.Close()

	return &data, nil
}
