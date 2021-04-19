package util

import (
	"encoding/json"
	"github.com/PaesslerAG/jsonpath"
	scraper "github.com/byung82/go-cloudflare-scraper"
	"github.com/levigross/grequests"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Company struct {
	Name              string      `json:"name"`
	CompanyNumber     string      `json:"company_number"`
	JurisdictionCode  string      `json:"jurisdiction_code"`
	IncorporationDate string      `json:"incorporation_date"`
	DissolutionDate   interface{} `json:"dissolution_date"`
	CompanyType       string      `json:"company_type"`
	RegistryURL       string      `json:"registry_url"`
	Branch            interface{} `json:"branch"`
	BranchStatus      interface{} `json:"branch_status"`
	Inactive          bool        `json:"inactive"`
	CurrentStatus     string      `json:"current_status"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	RetrievedAt       time.Time   `json:"retrieved_at"`
	OpencorporatesURL string      `json:"opencorporates_url"`
	Source            struct {
		Publisher   string    `json:"publisher"`
		URL         string    `json:"url"`
		Terms       string    `json:"terms"`
		RetrievedAt time.Time `json:"retrieved_at"`
	} `json:"source"`
	AgentName                      interface{}   `json:"agent_name"`
	AgentAddress                   interface{}   `json:"agent_address"`
	AlternativeNames               []interface{} `json:"alternative_names"`
	PreviousNames                  []interface{} `json:"previous_names"`
	NumberOfEmployees              interface{}   `json:"number_of_employees"`
	NativeCompanyNumber            interface{}   `json:"native_company_number"`
	AlternateRegistrationEntities  []interface{} `json:"alternate_registration_entities"`
	PreviousRegistrationEntities   []interface{} `json:"previous_registration_entities"`
	SubsequentRegistrationEntities []interface{} `json:"subsequent_registration_entities"`
	RegisteredAddressInFull        string        `json:"registered_address_in_full"`
	IndustryCodes                  []struct {
		IndustryCode struct {
			Code           string `json:"code"`
			Description    string `json:"description"`
			CodeSchemeID   string `json:"code_scheme_id"`
			CodeSchemeName string `json:"code_scheme_name"`
			UID            string `json:"uid"`
		} `json:"industry_code"`
	} `json:"industry_codes"`
	Identifiers            []interface{} `json:"identifiers"`
	TrademarkRegistrations []interface{} `json:"trademark_registrations"`
	RegisteredAddress      struct {
		StreetAddress string `json:"street_address"`
		Locality      string `json:"locality"`
		Region        string `json:"region"`
		PostalCode    string `json:"postal_code"`
		Country       string `json:"country"`
	} `json:"registered_address"`
	CorporateGroupings       []interface{} `json:"corporate_groupings"`
	Data                     interface{}   `json:"data"`
	FinancialSummary         interface{}   `json:"financial_summary"`
	HomeCompany              interface{}   `json:"home_company"`
	ControllingEntity        interface{}   `json:"controlling_entity"`
	UltimateBeneficialOwners []struct {
		UltimateBeneficialOwner struct {
			Name              string `json:"name"`
			OpencorporatesURL string `json:"opencorporates_url"`
		} `json:"ultimate_beneficial_owner"`
	} `json:"ultimate_beneficial_owners"`
	Filings []struct {
		Filing struct {
			ID                int         `json:"id"`
			Title             string      `json:"title"`
			Description       string      `json:"description"`
			UID               string      `json:"uid"`
			FilingTypeCode    string      `json:"filing_type_code"`
			FilingTypeName    string      `json:"filing_type_name"`
			URL               interface{} `json:"url"`
			OpencorporatesURL string      `json:"opencorporates_url"`
			Date              string      `json:"date"`
		} `json:"filing"`
	} `json:"filings"`
	Officers []struct {
		Officer struct {
			ID                int         `json:"id"`
			Name              string      `json:"name"`
			Position          string      `json:"position"`
			UID               interface{} `json:"uid"`
			StartDate         string      `json:"start_date"`
			EndDate           interface{} `json:"end_date"`
			OpencorporatesURL string      `json:"opencorporates_url"`
			Occupation        string      `json:"occupation"`
			Inactive          bool        `json:"inactive"`
			CurrentStatus     interface{} `json:"current_status"`
		} `json:"officer"`
	} `json:"officers"`
	Bag string
}

type OCResult struct {
	APIVersion string `json:"api_version"`
	Results    struct {
		Company Company `json:"company"`
	} `json:"results"`
}
type Pair struct {
	Key   string
	Value float64
}
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

var keys = map[string]int{
	"name": 0,
	//"description":    0,
	"company_number": 0,
	"wikipedia_id":   0,
	"street_address": 0,
	"locality":       0,
	"region":         0,
	"postal_code":    0,
	"country":        0,
	"industry_codes":        0,
}

func GetProxyRequestOptions() *grequests.RequestOptions {
	return nil
}

// REDACTED
//	return &grequests.RequestOptions{
//		Proxies: map[string]*url.URL{
//			"http":  proxyURL,
//			"https": proxyURL,
//		},
//		//UserAgent:
//	}
//}

// This is very bad and shouldn't be allowed to continue.
func GetCompanyKeywords(cid string) (*Company, error) {
	scraper, err := scraper.NewTransport(http.DefaultTransport)
	if err != nil {
		log.Fatal(err)
	}
	//
	c := http.Client{Transport: scraper}

	//r, err := grequests.Get("https://api.opencorporates.com/companies/gb/"+cid, GetProxyRequestOptions())
	r, err := c.Get("https://api.opencorporates.com/companies/gb/"+cid)
	if err != nil {
		logrus.WithError(err).Errorf("error getting '%s' from opencorporates", cid)
		return nil, err
	}
	var res *OCResult
	body, err := ioutil.ReadAll(r.Body)
	//println(string(body))
	if err = json.Unmarshal(body, &res); err == nil {
		res.Results.Company.Bag = JsonIterator(body)
		//logrus.WithField("company_id", cid).Infof("retrieved '%s' (%q) from opencorporates", res.Results.Company.Name, res.Results.Company.Bag)
		return &res.Results.Company, nil
	}
	return nil, err
}

func JsonIterator(j []byte) string {
	sb := strings.Builder{}
	var js map[string]interface{}
	err := json.Unmarshal(j, &js)
	if err != nil {
		logrus.WithError(err).Errorf("error decoding")
	}
	//json.Unmarshal([]byte(`{"api_version":"0.4.6","results":{"company":{"name":"CAPITA IT SERVICES LIMITED","company_number":"SC045439","jurisdiction_code":"gb","incorporation_date":"1968-02-07","dissolution_date":null,"company_type":"Private Limited Company","registry_url":"https://beta.companieshouse.gov.uk/company/SC045439","branch":null,"branch_status":null,"inactive":false,"current_status":"Active","created_at":"2010-10-21T20:03:36+00:00","updated_at":"2019-10-09T21:49:34+00:00","retrieved_at":"2019-10-09T21:49:30+00:00","opencorporates_url":"https://opencorporates.com/companies/gb/SC045439","source":{"publisher":"UK Companies House","url":"http://xmlgw.companieshouse.gov.uk/","terms":"UK Crown Copyright","retrieved_at":"2019-10-09T21:49:30+00:00"},"agent_name":null,"agent_address":null,"alternative_names":[],"previous_names":[{"company_name":"CARILLION IT SERVICES LIMITED","start_date":"2008-03-18","end_date":"2009-07-06"},{"company_name":"ALFRED MCALPINE - IT SERVICES LIMITED","start_date":"2007-04-04","end_date":"2008-03-18"},{"company_name":"ALFRED MCALPINE BUSINESS INFORMATION SYSTEMS LIMITED","start_date":"2004-05-10","end_date":"2007-04-04"},{"company_name":"MCALPINE BUSINESS INFORMATION SYSTEMS LIMITED","start_date":"2003-10-07","end_date":"2004-05-10"},{"company_name":"STIELL NETWORKS LIMITED","start_date":"1999-03-11","end_date":"2003-10-07"},{"company_name":"STIELL GLASGOW LIMITED","start_date":"1996-10-16","end_date":"1999-03-11"},{"company_name":"E.J. STIELL (GLASGOW) LIMITED","start_date":"1994-09-02","end_date":"1996-10-16"},{"company_name":"TOWN AND COUNTRY REFRIGERATION LIMITED.","start_date":"1968-02-07","end_date":"1994-09-02"}],"number_of_employees":null,"native_company_number":null,"alternate_registration_entities":[],"previous_registration_entities":[],"subsequent_registration_entities":[],"registered_address_in_full":"Pavilion Building Ellismuir Way\nTannochside Park, Uddingston, Glasgow, G71 5PW","industry_codes":[{"industry_code":{"code":"62.09","description":"Other information technology and computer service activities","code_scheme_id":"uk_sic_2007","code_scheme_name":"UK SIC Classification 2007","uid":"uk_sic_2007-6209"}},{"industry_code":{"code":"62.09","description":"Other information technology and computer service activities","code_scheme_id":"eu_nace_2","code_scheme_name":"European Community NACE Rev 2","uid":"eu_nace_2-6209"}},{"industry_code":{"code":"6209","description":"Other information technology and computer service activities","code_scheme_id":"isic_4","code_scheme_name":"UN ISIC Rev 4","uid":"isic_4-6209"}}],"identifiers":[],"trademark_registrations":[],"registered_address":{"street_address":"Pavilion Building Ellismuir Way\nTannochside Park","locality":"Uddingston","region":"Glasgow","postal_code":"G71 5PW","country":"United Kingdom"},"corporate_groupings":[{"corporate_grouping":{"name":"capita","wikipedia_id":"Capita_Group","opencorporates_url":"https://opencorporates.com/corporate_groupings/capita","updated_at":"2019-11-02T10:28:06+00:00"}}],"data":{"most_recent":[{"datum":{"id":1076378,"title":"VAT Number","data_type":"TaxNumber","description":"EU Vat Number: 618184140","opencorporates_url":"https://opencorporates.com/data/1076378"}},{"datum":{"id":4558822,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4558822"}},{"datum":{"id":4557904,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4557904"}},{"datum":{"id":4557903,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4557903"}},{"datum":{"id":4556672,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4556672"}},{"datum":{"id":4554017,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Hardware Maintenance","opencorporates_url":"https://opencorporates.com/data/4554017"}},{"datum":{"id":4553914,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4553914"}},{"datum":{"id":4553913,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4553913"}},{"datum":{"id":4552282,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4552282"}},{"datum":{"id":4552281,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4552281"}},{"datum":{"id":4552280,"title":"Financial Transaction","data_type":"FinancialTransaction","description":"Computer Hardware","opencorporates_url":"https://opencorporates.com/data/4552280"}},{"datum":{"id":3574931,"title":"Company Address","data_type":"CompanyAddress","description":"BIRCH STREET, WOLVERHAMPTON, WV1 4HY, United Kingdom","opencorporates_url":"https://opencorporates.com/data/3574931"}},{"datum":{"id":3574930,"title":"UK Data Protection Register entry","data_type":"OfficialRegisterEntry","description":"register id: Z9731197","opencorporates_url":"https://opencorporates.com/data/3574930"}},{"datum":{"id":1486775,"title":"Environmental Statement","data_type":"PublicStatement","description":"We recognise that our day to day activities impact on the environment in ways which are positive and negative. We wish to minimise these harmful effects wherever and whenever practicable, and will wor...","opencorporates_url":"https://opencorporates.com/data/1486775"}},{"datum":{"id":1486774,"title":"Description","data_type":"PublicStatement","description":"Capita IT Services provides IT Solutions for specific business needs. We know that there is no one solution that will suit every organisation. That is why, regardless of business size, or industry sec...","opencorporates_url":"https://opencorporates.com/data/1486774"}},{"datum":{"id":1076376,"title":"Approved UK Government Supplier","data_type":"GovernmentApprovedSupplier","description":null,"opencorporates_url":"https://opencorporates.com/data/1076376"}},{"datum":{"id":1076377,"title":"Trading Address","data_type":"CompanyAddress","description":"Ellismuir Way, Tannochside Business Park, Uddingston, Glasgow, Lanarkshire, G71 5PW","opencorporates_url":"https://opencorporates.com/data/1076377"}}],"total_count":37,"url":"https://opencorporates.com/companies/gb/SC045439/data"},"financial_summary":null,"home_company":null,"controlling_entity":null,"ultimate_beneficial_owners":[],"filings":[{"filing":{"id":696416855,"title":"Annual Accounts","description":"Full accounts made up to 2018-12-31","uid":"MzI0NTk1NjQxMGFkaXF6a2N4","filing_type_code":"AA","filing_type_name":"Annual Accounts","url":null,"opencorporates_url":"https://opencorporates.com/statements/696416855","date":"2019-10-08"}},{"filing":{"id":653492081,"title":"Confirmation Statement","description":"Confirmation statement made on 2019-06-30 with no updates","uid":"MzIzODkxMTk2NWFkaXF6a2N4","filing_type_code":"CS01","filing_type_name":"Confirmation Statement","url":null,"opencorporates_url":"https://opencorporates.com/statements/653492081","date":"2019-07-09"}},{"filing":{"id":650859890,"title":"Change of director's details","description":"Director's details changed for Mr John James Hemming on 2018-06-15","uid":"MzIzODYwMzYyN2FkaXF6a2N4","filing_type_code":"CH01","filing_type_name":"Change of director's details","url":null,"opencorporates_url":"https://opencorporates.com/statements/650859890","date":"2019-07-04"}},{"filing":{"id":650859893,"title":"Change of director's details","description":"Director's details changed for Mr Nicholas Siegfried Dale on 2018-06-15","uid":"MzIzODYwMzQzOGFkaXF6a2N4","filing_type_code":"CH01","filing_type_name":"Change of director's details","url":null,"opencorporates_url":"https://opencorporates.com/statements/650859893","date":"2019-07-04"}},{"filing":{"id":591695881,"title":"Annual Accounts","description":"Full accounts made up to 2017-12-31","uid":"MzIyNjMzMTg5MWFkaXF6a2N4","filing_type_code":"AA","filing_type_name":"Annual Accounts","url":null,"opencorporates_url":"https://opencorporates.com/statements/591695881","date":"2019-02-06"}},{"filing":{"id":574618137,"title":"Give notice of change of details for relevant legal entity with significant control","description":"Change of details for Capita It Services Holdings Limited as a person with significant control on 2018-10-01","uid":"MzIxNjIwOTQxMWFkaXF6a2N4","filing_type_code":"PSC05","filing_type_name":"Give notice of change of details for relevant legal entity with significant control","url":null,"opencorporates_url":"https://opencorporates.com/statements/574618137","date":"2018-10-04"}},{"filing":{"id":574618136,"title":"Change of corporate director's details","description":"Director's details changed for Capita Corporate Director Limited on 2018-06-15","uid":"MzIxNjIwOTUyOGFkaXF6a2N4","filing_type_code":"CH02","filing_type_name":"Change of corporate director's details","url":null,"opencorporates_url":"https://opencorporates.com/statements/574618136","date":"2018-10-04"}},{"filing":{"id":574618135,"title":"Change of corporate secretary's details","description":"Secretary's details changed for Capita Group Secretary Limited on 2018-06-15","uid":"MzIxNjIwOTUzMGFkaXF6a2N4","filing_type_code":"CH04","filing_type_name":"Change of corporate secretary's details","url":null,"opencorporates_url":"https://opencorporates.com/statements/574618135","date":"2018-10-04"}},{"filing":{"id":561245302,"title":"Confirmation Statement","description":"Confirmation statement made on 2018-06-30 with no updates","uid":"MzIwOTcwMjAwN2FkaXF6a2N4","filing_type_code":"CS01","filing_type_name":"Confirmation Statement","url":null,"opencorporates_url":"https://opencorporates.com/statements/561245302","date":"2018-07-13"}},{"filing":{"id":536423133,"title":"Appointment of director","description":"Appointment of Mr Nicholas Siegfried Dale as a director on 2018-01-31","uid":"MzE5NzMwMDQ5N2FkaXF6a2N4","filing_type_code":"AP01","filing_type_name":"Appointment of director","url":null,"opencorporates_url":"https://opencorporates.com/statements/536423133","date":"2018-02-08"}},{"filing":{"id":536423132,"title":"Termination of appointment of director ","description":"Termination of appointment of Ian Edward Jarvis as a director on 2018-01-31","uid":"MzE5NzMxMDYyNGFkaXF6a2N4","filing_type_code":"TM01","filing_type_name":"Termination of appointment of director ","url":null,"opencorporates_url":"https://opencorporates.com/statements/536423132","date":"2018-02-08"}},{"filing":{"id":529777157,"title":"Annual Accounts","description":"Full accounts made up to 2016-12-31","uid":"MzE5MzAzMDc2MWFkaXF6a2N4","filing_type_code":"AA","filing_type_name":"Annual Accounts","url":null,"opencorporates_url":"https://opencorporates.com/statements/529777157","date":"2017-12-18"}},{"filing":{"id":504418619,"title":"Appointment of director","description":"Appointment of Mr John James Hemming as a director on 2017-09-28","uid":"MzE4NjgwNzk4NWFkaXF6a2N4","filing_type_code":"AP01","filing_type_name":"Appointment of director","url":null,"opencorporates_url":"https://opencorporates.com/statements/504418619","date":"2017-10-02"}},{"filing":{"id":490548770,"title":"Confirmation Statement","description":"Confirmation statement made on 2017-06-30 with no updates","uid":"MzE4MDU5MzgyOGFkaXF6a2N4","filing_type_code":"CS01","filing_type_name":"Confirmation Statement","url":null,"opencorporates_url":"https://opencorporates.com/statements/490548770","date":"2017-07-14"}},{"filing":{"id":455038899,"title":"Termination of appointment of director ","description":"Termination of appointment of Richard John Shearer as a director on 2016-12-31","uid":"MzE2NTc0OTg4NmFkaXF6a2N4","filing_type_code":"TM01","filing_type_name":"Termination of appointment of director ","url":null,"opencorporates_url":"https://opencorporates.com/statements/455038899","date":"2017-01-03"}},{"filing":{"id":455038900,"title":"Appointment of director","description":"Appointment of Mr Ian Edward Jarvis as a director on 2016-12-21","uid":"MzE2NTczMDY2NWFkaXF6a2N4","filing_type_code":"AP01","filing_type_name":"Appointment of director","url":null,"opencorporates_url":"https://opencorporates.com/statements/455038900","date":"2017-01-03"}},{"filing":{"id":441274416,"title":"Annual Accounts","description":"Full accounts made up to 2015-12-31","uid":"MzE1OTUzMDYzOWFkaXF6a2N4","filing_type_code":"AA","filing_type_name":"Annual Accounts","url":null,"opencorporates_url":"https://opencorporates.com/statements/441274416","date":"2016-10-14"}},{"filing":{"id":393942909,"title":"Confirmation Statement","description":"Confirmation statement made on 2016-06-30 with updates","uid":"MzE1MjkxNTYxMGFkaXF6a2N4","filing_type_code":"CS01","filing_type_name":"Confirmation Statement","url":null,"opencorporates_url":"https://opencorporates.com/statements/393942909","date":"2016-07-14"}},{"filing":{"id":381702361,"title":"Termination of appointment of director ","description":"Termination of appointment of Lisa Ann Oxley as a director on 2016-05-18","uid":"MzE0OTEzMDUxOGFkaXF6a2N4","filing_type_code":"TM01","filing_type_name":"Termination of appointment of director ","url":null,"opencorporates_url":"https://opencorporates.com/statements/381702361","date":"2016-05-23"}},{"filing":{"id":379646395,"title":"Termination of appointment of director ","description":"Termination of appointment of Peter Hands as a director on 2016-04-30","uid":"MzE0ODEyNTAzMmFkaXF6a2N4","filing_type_code":"TM01","filing_type_name":"Termination of appointment of director ","url":null,"opencorporates_url":"https://opencorporates.com/statements/379646395","date":"2016-05-09"}},{"filing":{"id":206421269,"title":"Annual Accounts","description":"Full accounts made up to 2014-12-31","uid":"MzEzMjY2OTQzOWFkaXF6a2N4","filing_type_code":"AA","filing_type_name":"Annual Accounts","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421269","date":"2015-10-08"}},{"filing":{"id":206421271,"title":"Appointment of director","description":"Appointment of Miss Lisa Ann Oxley as a director on 2015-07-29","uid":"MzEyODA5ODAyMGFkaXF6a2N4","filing_type_code":"AP01","filing_type_name":"Appointment of director","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421271","date":"2015-07-30"}},{"filing":{"id":206421273,"title":"Annual Return","description":"Annual return made up to 2015-06-30 with full list of shareholders","uid":"MzEyNzkwNjkwOWFkaXF6a2N4","filing_type_code":"AR01","filing_type_name":"Annual Return","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421273","date":"2015-07-28"}},{"filing":{"id":206421274,"title":"Filing dated 2015-02-09","description":"","uid":"MzExNjk0Njk3MmFkaXF6a2N4","filing_type_code":"MISC","filing_type_name":null,"url":null,"opencorporates_url":"https://opencorporates.com/statements/206421274","date":"2015-02-09"}},{"filing":{"id":206421275,"title":"Filing dated 2014-12-18","description":"Auditor's resignation","uid":"MzExMzc4MjY0MmFkaXF6a2N4","filing_type_code":"AUD","filing_type_name":null,"url":null,"opencorporates_url":"https://opencorporates.com/statements/206421275","date":"2014-12-18"}},{"filing":{"id":206421277,"title":"Annual Accounts","description":"Full accounts made up to 2013-12-31","uid":"MzEwOTA2NjA2MmFkaXF6a2N4","filing_type_code":"AA","filing_type_name":"Annual Accounts","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421277","date":"2014-10-08"}},{"filing":{"id":206421278,"title":"Annual Return","description":"Annual return made up to 2014-06-30 with full list of shareholders","uid":"MzEwNDM1NTE3MGFkaXF6a2N4","filing_type_code":"AR01","filing_type_name":"Annual Return","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421278","date":"2014-07-24"}},{"filing":{"id":206421280,"title":"Change of director's details","description":"Director's details changed for Mr Richard John Shearer on 2013-09-12","uid":"MzA4OTU5NTQ1M2FkaXF6a2N4","filing_type_code":"CH01","filing_type_name":"Change of director's details","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421280","date":"2013-11-28"}},{"filing":{"id":206421282,"title":"Change of registered office address","description":"Registered office address changed from  Tannochside Park Uddingston Glasgow G71 5PW United Kingdom on 2013-10-01","uid":"MzA4NjEyMzUyOGFkaXF6a2N4","filing_type_code":"AD01","filing_type_name":"Change of registered office address","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421282","date":"2013-10-01"}},{"filing":{"id":206421283,"title":"Annual Accounts","description":"Full accounts made up to 2012-12-31","uid":"MzA4NTI3NTIxN2FkaXF6a2N4","filing_type_code":"AA","filing_type_name":"Annual Accounts","url":null,"opencorporates_url":"https://opencorporates.com/statements/206421283","date":"2013-09-18"}}],"officers":[{"officer":{"id":189497508,"name":"CAPITA GROUP SECRETARY LIMITED","position":"secretary","uid":null,"start_date":"2009-06-30","end_date":null,"opencorporates_url":"https://opencorporates.com/officers/189497508","occupation":null,"inactive":false,"current_status":null}},{"officer":{"id":189497529,"name":"JAMES DONALD","position":"secretary","uid":null,"start_date":"1992-02-12","end_date":"2000-04-07","opencorporates_url":"https://opencorporates.com/officers/189497529","occupation":null,"inactive":true,"current_status":null}},{"officer":{"id":189497537,"name":"GARRY JAMES FORSTER","position":"secretary","uid":null,"start_date":"2002-03-05","end_date":"2004-11-25","opencorporates_url":"https://opencorporates.com/officers/189497537","occupation":"CHARTERED SECRETARY","inactive":true,"current_status":null}},{"officer":{"id":189497544,"name":"CLAUDIA SUZANNE GOODMAN","position":"secretary","uid":null,"start_date":"2006-02-10","end_date":"2008-02-28","opencorporates_url":"https://opencorporates.com/officers/189497544","occupation":null,"inactive":true,"current_status":null}},{"officer":{"id":189497561,"name":"CAROLINE PATRICIA HIGGINS","position":"secretary","uid":null,"start_date":"2004-11-25","end_date":"2006-02-10","opencorporates_url":"https://opencorporates.com/officers/189497561","occupation":null,"inactive":true,"current_status":null}},{"officer":{"id":189497570,"name":"MACLAY MURRAY & SPENS LLP","position":"secretary","uid":null,"start_date":"2000-04-07","end_date":"2002-03-05","opencorporates_url":"https://opencorporates.com/officers/189497570","occupation":null,"inactive":true,"current_status":null}},{"officer":{"id":189497574,"name":"ALISON MARGARET SHEPLEY","position":"secretary","uid":null,"start_date":"2008-02-28","end_date":"2009-06-30","opencorporates_url":"https://opencorporates.com/officers/189497574","occupation":null,"inactive":true,"current_status":null}},{"officer":{"id":189497583,"name":"JOHN MURDOCH SINCLAIR","position":"secretary","uid":null,"start_date":"1989-06-28","end_date":"1992-02-12","opencorporates_url":"https://opencorporates.com/officers/189497583","occupation":"ACCOUNTANT","inactive":true,"current_status":null}},{"officer":{"id":189497597,"name":"RICHARD JOHN ADAM","position":"director","uid":null,"start_date":"2008-02-12","end_date":"2009-06-30","opencorporates_url":"https://opencorporates.com/officers/189497597","occupation":"FINANCE DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497610,"name":"WILLIAM MACDONALD ALLAN","position":"director","uid":null,"start_date":"1997-06-02","end_date":"2005-03-31","opencorporates_url":"https://opencorporates.com/officers/189497610","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497621,"name":"AM NOMINEES LIMITED","position":"director","uid":null,"start_date":"2006-02-01","end_date":"2008-02-12","opencorporates_url":"https://opencorporates.com/officers/189497621","occupation":"CORPORATE BODY","inactive":true,"current_status":null}},{"officer":{"id":189497633,"name":"JAMES DESMOND BARRETT","position":"director","uid":null,"start_date":"1989-06-28","end_date":"2002-10-11","opencorporates_url":"https://opencorporates.com/officers/189497633","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497643,"name":"CAPITA CORPORATE DIRECTOR LIMITED","position":"director","uid":null,"start_date":"2009-06-30","end_date":null,"opencorporates_url":"https://opencorporates.com/officers/189497643","occupation":null,"inactive":false,"current_status":null}},{"officer":{"id":189497655,"name":"STEPHEN PAUL CONNOR","position":"director","uid":null,"start_date":"2008-03-14","end_date":"2009-06-30","opencorporates_url":"https://opencorporates.com/officers/189497655","occupation":"OPERATIONS DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497661,"name":"JAMES DONALD","position":"director","uid":null,"start_date":"1993-05-31","end_date":"2000-04-07","opencorporates_url":"https://opencorporates.com/officers/189497661","occupation":"FINANCE DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497679,"name":"WILLIAM JAMES SPENCER FLOYDD","position":"director","uid":null,"start_date":"2012-10-11","end_date":"2013-06-17","opencorporates_url":"https://opencorporates.com/officers/189497679","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497690,"name":"PETER HANDS","position":"director","uid":null,"start_date":"2013-06-05","end_date":null,"opencorporates_url":"https://opencorporates.com/officers/189497690","occupation":"MANAGING DIRECTOR","inactive":false,"current_status":null}},{"officer":{"id":189497712,"name":"RODNEY HEWER HARRIS","position":"director","uid":null,"start_date":"2008-03-14","end_date":"2009-06-30","opencorporates_url":"https://opencorporates.com/officers/189497712","occupation":"FINANCE DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497725,"name":"LEE HEWETT","position":"director","uid":null,"start_date":"2009-06-30","end_date":"2011-06-10","opencorporates_url":"https://opencorporates.com/officers/189497725","occupation":"MANAGING DIRECTOR IT SERVICES","inactive":true,"current_status":null}},{"officer":{"id":189497738,"name":"ANDREW PHILIP JACKSON","position":"director","uid":null,"start_date":"2002-03-05","end_date":"2003-12-20","opencorporates_url":"https://opencorporates.com/officers/189497738","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497758,"name":"THOMAS DONALD KENNY","position":"director","uid":null,"start_date":"2008-03-14","end_date":"2009-06-30","opencorporates_url":"https://opencorporates.com/officers/189497758","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497767,"name":"DOMINIC JOSEPH LAVELLE","position":"director","uid":null,"start_date":"2004-12-06","end_date":"2007-04-23","opencorporates_url":"https://opencorporates.com/officers/189497767","occupation":"FINANCE DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497779,"name":"WILLIAM ALEXANDER LOCH","position":"director","uid":null,"start_date":"2000-07-01","end_date":"2004-10-01","opencorporates_url":"https://opencorporates.com/officers/189497779","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497801,"name":"STEWART MACKERRACHER","position":"director","uid":null,"start_date":"2003-01-01","end_date":"2006-05-19","opencorporates_url":"https://opencorporates.com/officers/189497801","occupation":"OPERATIONS DIR","inactive":true,"current_status":null}},{"officer":{"id":189497824,"name":"DAVID JOHNSTONE MCCALLUM","position":"director","uid":null,"start_date":"1989-06-28","end_date":"1997-06-20","opencorporates_url":"https://opencorporates.com/officers/189497824","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497834,"name":"JOHN MCDONOUGH","position":"director","uid":null,"start_date":"2008-02-12","end_date":"2009-06-30","opencorporates_url":"https://opencorporates.com/officers/189497834","occupation":"GROUP CHIEF EXECUTIVE","inactive":true,"current_status":null}},{"officer":{"id":189497846,"name":"CRAIG MATTHEW MCGILVRAY","position":"director","uid":null,"start_date":"2000-07-01","end_date":"2008-02-12","opencorporates_url":"https://opencorporates.com/officers/189497846","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497858,"name":"NOEL MCNULTY","position":"director","uid":null,"start_date":"2003-01-01","end_date":"2011-10-13","opencorporates_url":"https://opencorporates.com/officers/189497858","occupation":"FINANCE DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497875,"name":"RICHARD DAVID MOGG","position":"director","uid":null,"start_date":"2011-10-13","end_date":"2012-06-29","opencorporates_url":"https://opencorporates.com/officers/189497875","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497892,"name":"DOUGLAS JAMES MORE","position":"director","uid":null,"start_date":"2000-06-26","end_date":"2006-06-14","opencorporates_url":"https://opencorporates.com/officers/189497892","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497918,"name":"ANDREW GEORGE PARKER","position":"director","uid":null,"start_date":"2009-06-30","end_date":"2011-10-13","opencorporates_url":"https://opencorporates.com/officers/189497918","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497928,"name":"SIMON CHRISTOPHER PILLING","position":"director","uid":null,"start_date":"2009-06-30","end_date":"2011-01-10","opencorporates_url":"https://opencorporates.com/officers/189497928","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497944,"name":"ROGER WILLIAM ROBINSON","position":"director","uid":null,"start_date":"2008-02-12","end_date":"2009-06-30","opencorporates_url":"https://opencorporates.com/officers/189497944","occupation":"CIVIL ENGINEER","inactive":true,"current_status":null}},{"officer":{"id":189497967,"name":"JOHN MURDOCH SINCLAIR","position":"director","uid":null,"start_date":"1989-06-28","end_date":"1992-02-12","opencorporates_url":"https://opencorporates.com/officers/189497967","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497979,"name":"EDWARD GEORGE SMYTH","position":"director","uid":null,"start_date":"2000-05-03","end_date":"2004-04-30","opencorporates_url":"https://opencorporates.com/officers/189497979","occupation":"COMPANY DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":189497993,"name":"JOHN JAMES TAYLOR","position":"director","uid":null,"start_date":"2005-01-24","end_date":"2007-08-20","opencorporates_url":"https://opencorporates.com/officers/189497993","occupation":"CHARTERED ACCOUNTANT","inactive":true,"current_status":null}},{"officer":{"id":189498005,"name":"MARK RICHARD JOHN WYLLIE","position":"director","uid":null,"start_date":"2011-10-13","end_date":"2012-12-12","opencorporates_url":"https://opencorporates.com/officers/189498005","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":235166334,"name":"LISA ANN OXLEY","position":"director","uid":null,"start_date":"2015-07-29","end_date":"2016-05-18","opencorporates_url":"https://opencorporates.com/officers/235166334","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":255247297,"name":"RICHARD JOHN SHEARER","position":"director","uid":null,"start_date":"2013-06-04","end_date":"2016-12-31","opencorporates_url":"https://opencorporates.com/officers/255247297","occupation":"CHARTERED ACCOUNTANT","inactive":true,"current_status":null}},{"officer":{"id":255247298,"name":"IAN EDWARD JARVIS","position":"director","uid":null,"start_date":"2016-12-21","end_date":"2018-01-31","opencorporates_url":"https://opencorporates.com/officers/255247298","occupation":"DIRECTOR","inactive":true,"current_status":null}},{"officer":{"id":273469936,"name":"JOHN JAMES HEMMING","position":"director","uid":null,"start_date":"2017-09-28","end_date":null,"opencorporates_url":"https://opencorporates.com/officers/273469936","occupation":"DIRECTOR","inactive":false,"current_status":null}},{"officer":{"id":277907080,"name":"NICHOLAS SIEGFRIED DALE","position":"director","uid":null,"start_date":"2018-01-31","end_date":null,"opencorporates_url":"https://opencorporates.com/officers/277907080","occupation":"DIRECTOR","inactive":false,"current_status":null}}]}}}`), &js)
	for _, obj := range js {
		for k, _ := range keys {
			f, _ := jsonpath.Get("$.."+k, obj)
			for _, v := range f.([]interface{}) {
				if v != nil {
					switch v.(type) {
					case []interface{}:
						//fmt.Printf("%v", v.([]interface{}))
					case string:
						sb.WriteString(v.(string) + "\n")
					}
				}
			}
		}
	}
	//}
	return sb.String()
}

func AppendUniq(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}
