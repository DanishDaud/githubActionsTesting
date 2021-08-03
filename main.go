package main

import (
	// Println() function

	"fmt"
	"regexp"
	"strings"
)

func SimplifyPhoneNumber(source string) string {

	newString := strings.Replace(source, " ", "", -1)    // eliminate spaces
	newString = strings.Replace(newString, "-", "", -1)  // eliminate dashes
	newString = strings.Replace(newString, "(", "", -1)  // eliminate brackets
	newString = strings.Replace(newString, ")", "", -1)  // eliminate brackets
	newString = strings.Replace(newString, "/", "", -1)  // eliminate backslash
	newString = strings.Replace(newString, "\\", "", -1) // eliminate backslash
	newString = strings.Replace(newString, "\t", "", -1) // eliminate brackets

	return newString
}

func main() {

	m1 := regexp.MustCompile("^\\+?[0-9]+")
	//re := regexp.MustCompile(`[^aeiou]`)
	//fmt.Println(re.ReplaceAllStringFunc("seafood fool", strings.ToUpper))
	source := "6./50.2/50.68/90"
	r := regexp.MustCompile("[0-9+]+")
	a := r.FindAllString(source, -1)
	newString := strings.Join(a[:], "")
	fmt.Println(newString)
	fmt.Println(m1.FindAllString("*-/*-*/-*/+650+250.6890fdasfsdafasd-/**", -1))
	fmt.Println(m1.FindAllString("650+250.6890", -1))
	fmt.Println(m1.FindAllString("650 + 250.6890", -1))
	fmt.Println(SimplifyPhoneNumber("650.250.6890"))
	fmt.Println(SimplifyPhoneNumber("650.250.6890"))
	fmt.Println(SimplifyPhoneNumber("6.50.250.6890"))
	fmt.Println(SimplifyPhoneNumber("6.50.250.6890"))
	fmt.Println(SimplifyPhoneNumber("hi hello"))
	fmt.Println(SimplifyPhoneNumber("continous test"))
	fmt.Println(SimplifyPhoneNumber("final test"))

}

// func main() {

// 	number := "13124872852"
// 	match, err := regexp.MatchString("^(?:(?:\\+?1\\s*(?:[.-]\\s*)?)?(?:\\(\\s*([2-9]1[02-9]|[2-9][02-8]1|[2-9][02-8][02-9])\\s*\\)|([2-9]1[02-9]|[2-9][02-8]1|[2-9][02-8][02-9]))\\s*(?:[.-]\\s*)?)([2-9]1[02-9]|[2-9][02-9]1|[2-9][02-9]{2})\\s*(?:[.-]\\s*)?([0-9]{4})(?:\\s*(?:#|x\\.?|ext\\.?|extension)\\s*(\\d+))?|(anonymous)$", number)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Println(match)

// 	sNumber := SimplifyPhoneNumber(number)
// 	// here extract last 10 digits from the number
// 	last10 := number[len(sNumber)-10:]
// 	fmt.Println(last10)

// }
