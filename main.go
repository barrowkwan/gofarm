package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/barrowkwan/gofarm/devices"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/api"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

func main() {

	var (
		mcp_addresses  = []int{0x20}
		sprinklerGroup []*devices.SprinklerMCP23017Driver
		mainControl    int = 0
		mutex              = &sync.Mutex{}
	)

	sprinklerGroup = make([]*devices.SprinklerMCP23017Driver, len(mcp_addresses))
	fmt.Println(mcp_addresses)
	gbot := gobot.NewMaster()
	server := api.NewAPI(gbot)
	server.Port = "3000"
	server.Debug()
	server.Start()

	r := raspi.NewAdaptor()

	button := gpio.NewButtonDriver(r, "13")
	main_tap := gpio.NewLedDriver(r, "16")
	sensor := gpio.NewPIRMotionDriver(r, "15")
	//buzzer := gpio.NewBuzzerDriver(r, "18")

	main_tap.Off()

	for index, mcpAddr := range mcp_addresses {
		tmpIndex := index
		sprinklerGroup[tmpIndex] = devices.NewSprinklerMCP23017Driver(r, mcpAddr)

		for p := 0; p < 8; p++ {
			pin := p
			gbot.AddCommand("zone_"+strconv.Itoa(mcpAddr)+"_A_"+strconv.Itoa(pin)+"_on", func(params map[string]interface{}) interface{} {
				mutex.Lock()
				if sprinklerGroup[tmpIndex].ReadPin(uint8(pin), "A") == 0 {
					mainControl += 1
				}
				if !main_tap.State() {
					main_tap.On()
					time.Sleep(2 * time.Second)
				}
				sprinklerGroup[tmpIndex].On("A", uint8(pin))
				mutex.Unlock()
				return "zone_" + strconv.Itoa(mcpAddr) + "_A_" + strconv.Itoa(pin) + "_on"
			})
			gbot.AddCommand("zone_"+strconv.Itoa(mcpAddr)+"_A_"+strconv.Itoa(pin)+"_off", func(params map[string]interface{}) interface{} {
				mutex.Lock()
				if sprinklerGroup[tmpIndex].ReadPin(uint8(pin), "A") == (1 << uint8(pin)) {
					mainControl -= 1
				}
				sprinklerGroup[tmpIndex].Off("A", uint8(pin))
				if mainControl == 0 {
					main_tap.Off()
					time.Sleep(2 * time.Second)
				}
				mutex.Unlock()
				return "zone_" + strconv.Itoa(mcpAddr) + "_A_" + strconv.Itoa(pin) + "_off"
			})
			gbot.AddCommand("zone_"+strconv.Itoa(mcpAddr)+"_B_"+strconv.Itoa(pin)+"_on", func(params map[string]interface{}) interface{} {
				mutex.Lock()
				if sprinklerGroup[tmpIndex].ReadPin(uint8(pin), "B") == 0 {
					mainControl += 1
				}
				if !main_tap.State() {
					main_tap.On()
					time.Sleep(2 * time.Second)
				}
				sprinklerGroup[tmpIndex].On("B", uint8(pin))
				mutex.Unlock()
				return "zone_" + strconv.Itoa(mcpAddr) + "_B_" + strconv.Itoa(pin) + "_on"
			})
			gbot.AddCommand("zone_"+strconv.Itoa(mcpAddr)+"_B_"+strconv.Itoa(pin)+"_off", func(params map[string]interface{}) interface{} {
				mutex.Lock()
				if sprinklerGroup[tmpIndex].ReadPin(uint8(pin), "B") == (1 << uint8(pin)) {
					mainControl -= 1
				}
				sprinklerGroup[tmpIndex].Off("B", uint8(pin))
				if mainControl == 0 {
					main_tap.Off()
					time.Sleep(2 * time.Second)
				}
				mutex.Unlock()
				return "zone_" + strconv.Itoa(mcpAddr) + "_B_" + strconv.Itoa(pin) + "_off"
			})
		}
	}

	gbot.AddCommand("main_on", func(params map[string]interface{}) interface{} {
		main_tap.On()
		return "Main Water Tap is ON"
	})

	gbot.AddCommand("main_off", func(params map[string]interface{}) interface{} {
		main_tap.Off()
		return "Main Water Tap is OFF"
	})

	work := func() {

		button.On(gpio.ButtonPush, func(data interface{}) {
			fmt.Println("button pressed")
			main_tap.On()
		})

		button.On(gpio.ButtonRelease, func(data interface{}) {
			fmt.Println("button released")
			main_tap.Off()
		})
		sensor.On(gpio.MotionDetected, func(data interface{}) {
			fmt.Println(gpio.MotionDetected)
			//buzzerOn(buzzer)
			//main_tap.On()
		})
		sensor.On(gpio.MotionStopped, func(data interface{}) {
			fmt.Println(gpio.MotionStopped)
			//main_tap.Off()
		})
	}

	robot := gobot.NewRobot("GoframBot",
		[]gobot.Connection{r},
		[]gobot.Device{button, main_tap, sensor, sprinklerGroup[0]},
		work,
	)

	robot = gbot.AddRobot(robot)
	robot.AddCommand("say_hello", func(params map[string]interface{}) interface{} {
		fmt.Println(params)
		return fmt.Sprintf("%v says hello!", robot.Name)
	})

	gbot.Start()

}

func buzzerOn(b *gpio.BuzzerDriver) {

	type note struct {
		tone     float64
		duration float64
	}

	song := []note{
		{gpio.C4, gpio.Quarter},
		{gpio.C4, gpio.Quarter},
		{gpio.G4, gpio.Quarter},
		{gpio.G4, gpio.Quarter},
		{gpio.A4, gpio.Quarter},
		{gpio.A4, gpio.Quarter},
		{gpio.G4, gpio.Half},
		{gpio.F4, gpio.Quarter},
		{gpio.F4, gpio.Quarter},
		{gpio.E4, gpio.Quarter},
		{gpio.E4, gpio.Quarter},
		{gpio.D4, gpio.Quarter},
		{gpio.D4, gpio.Quarter},
		{gpio.C4, gpio.Half},
	}

	for _, val := range song {
		b.Tone(val.tone, val.duration)
		time.Sleep(10 * time.Millisecond)
	}
}
