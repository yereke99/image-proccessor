package tests

import (
	"ImageProcessor/config"
	"ImageProcessor/pipeline"
	"fmt"
	"github.com/gogf/gf/v2/text/gstr"
	"log"
	"regexp"
	"strings"
	"testing"
)

var v []int

func f(a []int) []int {
	a = append(a, 2, 3, 4)
	return a
}
func TestEd(t *testing.T) {
	v = make([]int, 0, 4)

	v = append(v, 1)
	f(v)

	fmt.Println(v[:4])
	fmt.Println(v)
	fmt.Println(v[2])
}

func TestNewPull(t *testing.T) {
	//takeImageInspector(space, from, to, priority)
	fmt.Print("-> Create configuration...")
	conf, err := config.NewConfig("3.134.16.137:4470")
	if err != nil {
		log.Fatal("Conf broken: ", err)
	}

	resp, err := conf.Taran.Conn.Call("takeImageInspector", []interface{}{"DCCache", "s", "testme", "LOW"})
	if err != nil {
		log.Println("*** Error when pull job:", err.Error())
		log.Fatal(err)
	}
	log.Println(resp)

}

func TestLeven(t *testing.T) {
	text := `"6:07  ca)  SY  +1 (231) 241-2514  Pentwater, MI   ]  @  Remind Me  Message "`
	num := `12312412514`

	reg := regexp.MustCompile(`([^0-9])`)
	newText := reg.ReplaceAllString(text, "")

	result := gstr.Levenshtein(num, newText, 1, 1, 1)

	if result <= 7 {
		fmt.Println("YES")
	}

	number := num[1:]
	if strings.Contains(newText, number) {
		fmt.Println("YES")
	}
}

func TestCnam(t *testing.T) {

	text := "Salona So v  CSS itsyabr |  © Incoming call  Oclinicals/Digestive Health  Dallas, 1X  an  Block/report number"
	cnam := "Oclinicals"

	fmt.Println(pipeline.CheckCnam(text, cnam))

}

func TestSmartSplit(t *testing.T) {
	text := `
 Incoming number match FALSE
 text 6094. KA * WC KH |
 Incoming call
 1-800Accountant
 n
 Number Verified
 Block/report number
`
	re := regexp.MustCompile(`[^A-z]`)
	split := re.Split(text, -1)
	for _, v := range split {
		if v == "" {
			continue
		}
		fmt.Println(strings.TrimSpace(v))
	}

}

func TestNumber(t *testing.T) {

	var task pipeline.Task

	task.Id = 1
	task.Fname = ""
	//task.FromNum = "15738762600"
	task.FromNum = "13182428103"
	task.MessageId = "d07cc57b-002c-4f8d-8d93-3947e349a0d9"
	//task.RecognitedText = "all S\\Veterans United\\Columbia, MO\\na a\\nRemind Me Message\\nDecline Accept"
	task.RecognitedText = "a Verizon & MA) [o2) +1 (404) 595-3108 Atlanta, GA a @ iecraal are myc) Message answe @"
	task.ImageBytes = []byte("")
	task.NumberMatch = false

	count := 0
	numArray := strings.Split(task.FromNum, "")

	reg := regexp.MustCompile(`([\d]{2,3}%)`)
	newText := reg.ReplaceAllString(task.RecognitedText, "")

	reg = regexp.MustCompile(`[\d]{1,2}:[\d]{1,2}`)
	newText = reg.ReplaceAllString(newText, "")

	reg = regexp.MustCompile(`([^0-9])`)
	newText = reg.ReplaceAllString(newText, "")
	fmt.Println("New text: ", newText)

	if len(newText) < 8 {
		t.Fatal("false")
	}

	for _, digit := range numArray {

		index := strings.Index(newText, digit)
		if index >= 0 {
			count = count + 1
			newText = strings.Replace(newText, digit, "", 1)
		}
	}

	fmt.Println("count", count)

	if count < 8 {
		t.Fatal("false")
	}

	t.Fatal("true")
}

func TestLivenstein(t *testing.T) {

	fWord := "14045953056"
	//text := "a Verizon & MA) [o2) +1 (404) 595-3108 Atlanta, GA a @ iecraal are myc) Message answe @"
	//text := "Pre 0) 50] (ie  ee |  [Om 52)  +1 (646) 647-8343  New York, NY  )  @  iecraal are myc)  Message  an,  Decline  Accept"
	text := "Pre 0) 50] (ie  ee ew  (o2)  +1 (404) 595-3056  Atlanta, GA  )  @  iecraal are myc)  Message  an,  Decline  Accep"
	reg := regexp.MustCompile(`([^0-9])`)
	text = reg.ReplaceAllString(text, "")

	fmt.Println("New text", text)
	result := gstr.Levenshtein(fWord, text, 1, 1, 1)
	fmt.Println("result", result)
	if result < 5 {
		return
	}
}
