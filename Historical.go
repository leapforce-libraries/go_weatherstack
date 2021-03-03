package weatherstack

import (
	"fmt"
	"net/url"
	"time"

	"cloud.google.com/go/civil"
	errortools "github.com/leapforce-libraries/go_errortools"
	go_http "github.com/leapforce-libraries/go_http"
	utilities "github.com/leapforce-libraries/go_utilities"
)

type Hourly int

const (
	HourlyOn  Hourly = 1
	HourlyOff Hourly = 0
)

type Interval int

const (
	Interval1Hour      Interval = 1
	Interval3Hours     Interval = 3
	Interval6Hours     Interval = 6
	IntervalDayNight   Interval = 12
	IntervalDayAverage Interval = 24
)

type Units string

const (
	UnitsMetric     Units = "m"
	UnitsScientific Units = "s"
	UnitsFahrenheit Units = "f"
)

type HistoricalResponse struct {
	Request    Request                      `json:"request"`
	Location   Location                     `json:"location"`
	Current    CurrentWeather               `json:"current"`
	Historical map[string]HistoricalWeather `json:"historical"`
}

type Request struct {
	Type     string `json:"type"`
	Query    string `json:"query"`
	Language string `json:"language"`
	Unit     string `json:"unit"`
}

type Location struct {
	Name           string `json:"name"`
	Country        string `json:"country"`
	Region         string `json:"region"`
	Lat            string `json:"lat"`
	Lon            string `json:"lon"`
	TimezoneID     string `json:"timezone_id"`
	Localtime      string `json:"localtime"`
	LocaltimeEpoch int64  `json:"localtime_epoch"`
	UTCOffset      string `json:"utc_offset"`
}

type CurrentWeather struct {
	ObservationTime     string   `json:"observation_time"`
	Temperature         int      `json:"temperature"`
	WeatherCode         int      `json:"weather_code"`
	WeatherIcons        []string `json:"weather_icons"`
	WeatherDescriptions []string `json:"weather_descriptions"`
	WindSpeed           int      `json:"wind_speed"`
	WindDegree          int      `json:"wind_degree"`
	WindDir             string   `json:"wind_dir"`
	Pressure            int      `json:"pressure"`
	Precip              float64  `json:"precip"`
	Humidity            int      `json:"humidity"`
	Cloudcover          int      `json:"cloudcover"`
	FeelsLike           int      `json:"feelslike"`
	UVIndex             int      `json:"uv_index"`
	Visibility          int      `json:"visibility"`
	IsDay               string   `json:"is_day"`
}

type HistoricalWeather struct {
	Date      string          `json:"date"`
	DateEpoch int64           `json:"date_epoch"`
	Astro     Astro           `json:"astro"`
	MinTemp   int             `json:"mintemp"`
	MaxTemp   int             `json:"maxtemp"`
	AvgTemp   int             `json:"avgtemp"`
	TotalSnow float64         `json:"totalsnow"`
	SunHour   float64         `json:"sunhour"`
	UVIndex   int             `json:"uv_index"`
	Hourly    []HourlyWeather `json:"hourly"`
}

type Astro struct {
	Sunrise          string `json:"sunrise"`
	Sunset           string `json:"sunset"`
	Moonrise         string `json:"moonrise"`
	Moonset          string `json:"moonset"`
	MoonPhase        string `json:"moon_phase"`
	MoonIllumination int    `json:"moon_illumination"`
}

type HourlyWeather struct {
	Time                string   `json:"time"`
	Temperature         int      `json:"temperature"`
	WindSpeed           int      `json:"wind_speed"`
	WindDegree          int      `json:"wind_degree"`
	WindDir             string   `json:"wind_dir"`
	WeatherCode         int      `json:"weather_code"`
	WeatherIcons        []string `json:"weather_icons"`
	WeatherDescriptions []string `json:"weather_descriptions"`
	Precip              float64  `json:"precip"`
	Humidity            int      `json:"humidity"`
	Visibility          int      `json:"visibility"`
	Pressure            int      `json:"pressure"`
	Cloudcover          int      `json:"cloudcover"`
	Heatindex           int      `json:"heatindex"`
	Dewpoint            int      `json:"dewpoint"`
	Windchill           int      `json:"windchill"`
	Windgust            int      `json:"windgust"`
	FeelsLike           int      `json:"feelslike"`
	ChanceOfRain        int      `json:"chanceofrain"`
	ChanceOfRemDry      int      `json:"chanceofremdry"`
	ChanceOfWindy       int      `json:"chanceofwindy"`
	ChanceOfOvercast    int      `json:"chanceofovercast"`
	ChanceOfSunshine    int      `json:"chanceofsunshine"`
	ChanceOfFrost       int      `json:"chanceoffrost"`
	ChanceOfHighTemp    int      `json:"chanceofhightemp"`
	ChanceOfFog         int      `json:"chanceoffog"`
	ChanceOfSnow        int      `json:"chanceofsnow"`
	ChanceOfThunder     int      `json:"chanceofthunder"`
	UVIndex             int      `json:"uv_index"`
}

type GetHistoricalWeatherConfig struct {
	Query     string
	StartDate civil.Date
	EndDate   *civil.Date
	Hourly    *Hourly
	Interval  *Interval
	Units     *Units
	Language  *string
}

func (service *Service) GetHistoricalWeather(config GetHistoricalWeatherConfig) (*HistoricalResponse, *errortools.Error) {
	values := url.Values{}

	startDate := utilities.DateToTime(config.StartDate)

	if config.EndDate == nil {
		values.Add("historical_date", startDate.Format(DateFormat))
	} else {
		endDate := utilities.DateToTime(*config.EndDate)

		if startDate.After(endDate) {
			return nil, errortools.ErrorMessage("StartDate must be smaller or equal to EndDate.")
		}

		maxEndDate := startDate.Add(time.Duration(MaxDaysPerCall-1) * 24 * time.Hour)

		if endDate.After(maxEndDate) {
			return nil, errortools.ErrorMessage("Maximum time frame of 60 days exceeded.")
		}

		values.Add("historical_date_start", startDate.Format(DateFormat))
		values.Add("historical_date_end", endDate.Format(DateFormat))
	}

	values.Add("query", config.Query)

	if config.Hourly != nil {
		values.Add("hourly", fmt.Sprintf("%v", int(*config.Hourly)))
	}

	if config.Interval != nil {
		values.Add("interval", fmt.Sprintf("%v", int(*config.Interval)))
	}

	if config.Units != nil {
		values.Add("units", fmt.Sprintf("%s", string(*config.Units)))
	}

	if config.Language != nil {
		values.Add("language", *config.Language)
	}

	historicalResponse := HistoricalResponse{}

	requestConfig := go_http.RequestConfig{
		URL:           service.url(fmt.Sprintf("%s?%s", "historical", values.Encode())),
		ResponseModel: &historicalResponse,
	}

	_, _, e := service.get(&requestConfig)
	if e != nil {
		return nil, e
	}

	return &historicalResponse, nil
}
