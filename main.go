package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/sclevine/agouti"
)

func main() {
	poll()
}

func poll() {
	interval := time.Duration(getEnv("VANGUARD_INTERVAL", 10)) * time.Second
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions(
			"args", []string{
				"--headless",
				"--disable-gpu",
			},
		),
	)
	if err := driver.Start(); err != nil {
		panic(err)
	}

	page, err := driver.NewPage()
	if err != nil {
		panic(err)
	}

	login(page)
	for {
		fmt.Println("Vanguard.....")
		percentChanged, err := getPercentageChange(page)
		if err != nil {
			panic(err)
		}

		totalValue, err := getTotalValue(page)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Current value: %s, Percent change: %s", totalValue, percentChanged)

		page.Refresh()
		time.Sleep(interval)
		clearScreen()
	}
	driver.Stop()
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func login(page *agouti.Page) {
	vanguardUser := os.Getenv("VANGUARD_USERNAME")
	vanguardPassword := os.Getenv("VANGUARD_PASSWORD")

	if err := page.Navigate("https://secure.vanguardinvestor.co.uk/Login"); err != nil {
		panic(err)
	}

	const userInputId = "#__GUID_1006"
	const passwordInputId = "#__GUID_1007"
	const loginButtonClass = "button.btn-primary"
	userInput := page.Find(userInputId)
	passwordInput := page.Find(passwordInputId)
	userInput.Fill(vanguardUser)
	passwordInput.Fill(vanguardPassword)
	loginBtn := page.Find(loginButtonClass)
	loginBtn.Click()
}

func getPercentageChange(page *agouti.Page) (string, error) {
	return getText(page, "span.value-l")
}

func getTotalValue(page *agouti.Page) (string, error) {
	return getText(page, "div.value")
}

func getText(page *agouti.Page, textToFind string) (string, error) {
	maxAttempts := getEnv("VANGUARD_MAX_ATTEMPTS", 20)
	attempts := 0
	waitTime := time.Duration(getEnv("VANGUARD_WAIT_TIME", 1)) * time.Second
	for {
		found, _ := page.Find(textToFind).Text()
		if len(found) > 0 {
			return found, nil
		}
		if attempts == maxAttempts {
			return "", fmt.Errorf("Max attempts exceed. Cannot find %q", textToFind)
		}
		time.Sleep(waitTime)
		attempts += 1
	}
}

func getEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return intVal
	}
	return fallback
}
