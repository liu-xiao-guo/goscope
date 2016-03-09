package main

import (
	"launchpad.net/go-unityscopes/v2"
	"log"
	"encoding/json"
	"net/url"
	"net/http"
)


const searchCategoryYellow = `{
    "schema-version" : 1,
    "template" : {
        "category-layout" : "vertical-journal",
        "card-layout": "horizontal",
        "card-size": "small",
        "collapsed-rows": 0
     },
    "components" : {
        "title" : "title",
        "subtitle":"subtitle",
        "summary":"summary",
        "art":{
        	"field": "art",
       		"aspect-ratio": 1
        }
    }
}`

const searchCategoryTemplate = `{
    "schema-version" : 1,  
    "template" : {  
        "category-layout" : "carousel",  
        "card-size": "large",  
        "overlay" : true  
    },  
    "components" : {  
        "title" : "title",  
        "art" : {  
            "field": "art",  
            "aspect-ratio": 1.6,  
            "fill-mode": "fit"  
        }  
    }  
}`

// SCOPE ***********************************************************************

var scope_interface scopes.Scope

type MyScope struct {
	BaseURI string
	Key     string
	URI		string
	Dir     string	
	base *scopes.ScopeBase
}

type WathereResponse struct {
	WeatherList []Weather `json:"results"`
	Date string `json:"date"`
}

type Weather struct {
	CurrentCity string `json:"currentCity"`
	Pm25 string `json:"pm25"`
	IndexList []Index `json:"index"`
	Weather_datalist []Weather_data `json:"weather_data"`
}

type Index struct {
	Title string `json:"title"`
	Zs string `json:"zs"`
	Tipt string `json:"tipt"`
	Des string `json:"des"`
}

type Weather_data struct {
	Date string `json:"date"`
	DayPictureUrl string `json:"dayPictureUrl"`
	NightPictureUrl string `json:"nightPictureUrl"`
	Weather string `json:"weather"`
	Wind string `json:"wind"`
	Temperature string `json:"temperature"`
}

func (s *MyScope) buildUrl(url2 string, params map[string]string) string {
	query := make(url.Values)
	for key, value := range params {
		query.Set(key, value)
	}
	log.Println(url2 + query.Encode())
	return url2 + query.Encode()
}

// This is used to get results from a webservice
func (s *MyScope) get(url string, params map[string]string, result interface{}) error {
	resp, err := http.Get(s.buildUrl(url, params))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(result)
}

func (s *MyScope) Search(q *scopes.CannedQuery, metadata *scopes.SearchMetadata, reply *scopes.SearchReply, cancelled <-chan bool) error {
	root_department := s.CreateDepartments(q, metadata, reply)
	reply.RegisterDepartments(root_department)

	query := q.QueryString()
	log.Println(query)	
	
	// Try to get the city name
	loc := metadata.Location()
	city := loc.City;
	log.Println("city: ", city)
	
	if query == "" {
		if q.DepartmentID() == "" {
			query = city
		} else {
			query = q.DepartmentID()
		}			
	} 
	
	log.Println("query: ", query)
	
	var response WathereResponse
	
	if err := s.get(s.BaseURI, map[string]string{"location": query, "ak": s.Key, "output": "json"}, &response); err != nil {
		return err
	} else {
		log.Println("there is no error!")
	}
	
	// log.Println(response)
	date := response.Date;
	log.Println("date: ", date)
	
	var cat *scopes.Category;
	
	if len(q.QueryString()) == 0 && q.DepartmentID() == "" {
		cat = reply.RegisterCategory("weather", query, "", searchCategoryTemplate)
	} else {
		cat = reply.RegisterCategory("weather", "", "", searchCategoryTemplate)
	}
		
	
	for _, data := range response.WeatherList {
		result := scopes.NewCategorisedResult(cat)
		result.SetURI(s.URI)

//		log.Println("Current city:", data.CurrentCity)
//		log.Println("PM25: ", data.Pm25)
		
		var yellocalendar string = ""
		for _, index := range data.IndexList {
//			log.Println("title: ", index.Title)
//			log.Println("zs: ", index.Zs)
//			log.Println("tipt: ", index.Tipt)
//			log.Println("Des: ", index.Des)
			
			yellocalendar += index.Title + " "
			yellocalendar += index.Zs + " "
			yellocalendar += index.Tipt + " "
			yellocalendar += index.Des
		}
		
		for i, weather := range data.Weather_datalist {
//			log.Println("date: ", weather.Date)
//			log.Println("dayPictureUrl: ", weather.DayPictureUrl)
//			log.Println("nightPictureUrl: ", weather.NightPictureUrl)
//			log.Println("weather: ",weather.Weather)
//			log.Println("wind: ", weather.Wind)
//			log.Println("temperature: ", weather.Temperature)
			
			result.SetTitle(weather.Date)
			result.SetArt(weather.DayPictureUrl)
			result.Set("wind", weather.Wind)
			result.Set("weather", weather.Weather)
			result.Set("temperature", weather.Temperature)
			
			if err := reply.Push(result); err != nil {
				return err
			}
			
			result.SetArt(weather.NightPictureUrl)
			if err := reply.Push(result); err != nil {
				return err
			}			
			
			// Push the yellow calender now
			if i == 0  {
				cat1 := reply.RegisterCategory("weather1", "今天天气", "", searchCategoryYellow)
				result1 := scopes.NewCategorisedResult(cat1)								
				
				result1.SetURI(s.URI)
				result1.SetTitle(date)
				result1.SetArt(weather.DayPictureUrl)
				result1.Set("subtitle", weather.Weather + " " + weather.Wind + " " +  weather.Temperature + "  PMI: " +  data.Pm25)
				result1.Set("summary", yellocalendar)
				
				if err := reply.Push(result1); err != nil {
					return err
				}					
			}		
		}		
	}
	
	return nil
}

func (s *MyScope) Preview(result *scopes.Result, metadata *scopes.ActionMetadata, reply *scopes.PreviewReply, cancelled <-chan bool) error {
	layout1col := scopes.NewColumnLayout(1)
	layout2col := scopes.NewColumnLayout(2)
	layout3col := scopes.NewColumnLayout(3)

	// Single cyolumn layout
	layout1col.AddColumn("header", "image",  "wind", "weather", "temperature", "summary", "actions")

	// Two column layout
	layout2col.AddColumn("header")
	layout2col.AddColumn("image", "wind", "weather", "temperature", "summary", "actions")

	// Three cokumn layout
	layout3col.AddColumn("header")
	layout3col.AddColumn("image", "wind", "weather", "temperature","summary", "actions")
	layout3col.AddColumn()

	// Register the layouts we just created
	reply.RegisterLayout(layout1col, layout2col, layout3col)

	header := scopes.NewPreviewWidget("header", "header")

	// It has title and a subtitle properties
	header.AddAttributeMapping("title", "title")
	header.AddAttributeMapping("subtitle", "subtitle")

	// Define the image section
	image := scopes.NewPreviewWidget("image", "image")
	// It has a single source property, mapped to the result's art property
	image.AddAttributeMapping("source", "art")

	// Define the summary section
	description := scopes.NewPreviewWidget("summary", "text")
	description.AddAttributeMapping("text", "summary")

	wind := scopes.NewPreviewWidget("wind", "text")
	wind.AddAttributeMapping("text", "wind")

	weather := scopes.NewPreviewWidget("weather", "text")
	weather.AddAttributeMapping("text", "weather")

	temperature := scopes.NewPreviewWidget("temperature", "text")
	temperature.AddAttributeMapping("text", "temperature")

	// build variant map.
	var uri string

	if err := result.Get("uri", &uri); err != nil {
		log.Println(err)
	}

	tuple1 := make(map[string]interface{})
	tuple1["id"] = "open"
	tuple1["label"] = "Open"
	tuple1["uri"] = uri

	actions := scopes.NewPreviewWidget("actions", "actions")
	actions.AddAttributeValue("actions", []interface{}{tuple1})

	var summary string
	if err := result.Get("summary", &summary); err != nil {
		log.Println(err)
	}

	if len(summary) > 0 {
		reply.PushWidgets(header, image, description, actions)
	} else {
		reply.PushWidgets(header, image, wind, weather, temperature, actions)
	}

	return nil
}

func (s *MyScope) SetScopeBase(base *scopes.ScopeBase) {
	s.base = base
}


func (s *MyScope) GetSubdepartments1(query *scopes.CannedQuery,
	metadata *scopes.SearchMetadata,
	reply *scopes.SearchReply) *scopes.Department {
	active_dep, err := scopes.NewDepartment("wuhan", query, "湖北")
	
	if err == nil {
		department, _ := scopes.NewDepartment("武汉", query, "武汉")
		active_dep.AddSubdepartment(department)

		department2, _ := scopes.NewDepartment("宜昌", query, "宜昌")
		active_dep.AddSubdepartment(department2)
				
		department3, _ := scopes.NewDepartment("随州", query, "随州")
		active_dep.AddSubdepartment(department3)
	}

	return active_dep
}

func (s *MyScope) GetSubdepartments2(query *scopes.CannedQuery,
	metadata *scopes.SearchMetadata,
	reply *scopes.SearchReply) *scopes.Department {
	active_dep, err := scopes.NewDepartment("changsha", query, "湖南")
	
	if err == nil {
		department, _ := scopes.NewDepartment("长沙", query, "长沙")
		active_dep.AddSubdepartment(department)

		department2, _ := scopes.NewDepartment("株洲", query, "株洲")
		active_dep.AddSubdepartment(department2)
	}

	return active_dep
}

func (s *MyScope) CreateDepartments(query *scopes.CannedQuery,
	metadata *scopes.SearchMetadata,
	reply *scopes.SearchReply) *scopes.Department {
		
	department, _ := scopes.NewDepartment("", query, "选择地点")
//	department.SetAlternateLabel("Browse Music Alt")

	dept1 := s.GetSubdepartments1(query, metadata, reply)
	if dept1 != nil {
		department.AddSubdepartment(dept1)
	}

	dept2 := s.GetSubdepartments2(query, metadata, reply)
	if dept2 != nil {
		department.AddSubdepartment(dept2)
	}

	return department
}

// MAIN ************************************************************************

func main() {
	scope := &MyScope {
		BaseURI: "http://api.map.baidu.com/telematics/v3/weather?",
		Key:     "DdzwVcsGMoYpeg5xQlAFrXQt",
		URI: 	 "http://www.weather.com.cn/html/weather/101010100.shtml",
		Dir:     "",
	}
	
//	var sc MyScope
	scope_interface = scope
	
	if err := scopes.Run(scope); err != nil {
		log.Fatalln(err)
	}
}
