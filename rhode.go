package main

import (
	"fmt"
	"os"
	"sort"
	"time"
	"os/exec"
	"io/ioutil"
	"gocv.io/x/gocv"
	"strings"
)

func PathExists(path string, makedir bool) bool {
	isExist := false
	_, err := os.Stat(path)
	if err == nil {
		isExist = true
	}
	if os.IsNotExist(err) {
		isExist = false
	}
	if !isExist && makedir {
		os.Mkdir(path, os.ModePerm)
	}
	return isExist
}

func GetAllFile(pathname string, s []string, ext string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			s, err = GetAllFile(fullDir, s, ext)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return s, err
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			if strings.HasSuffix(fullName, ext) {
				s = append(s, fullName)
			}
		}
	}
	return s, nil
}

func Typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func execute(cmd string, args []string) {
	fmt.Println(fmt.Sprintf("%+v %+v", cmd, strings.Join(args, " ")))
	exec.Command(cmd, args...)
}

func findPos(template string, target string) []int {
	target_gray := gocv.IMRead(target, gocv.IMReadGrayScale)
	template_gray := gocv.IMRead(template, gocv.IMReadGrayScale)
	t_h, t_w := target_gray.Size()[0], target_gray.Size()[1]
	ret := gocv.NewMat()
	defer ret.Close()
	gocv.MatchTemplate(target_gray, template_gray, &ret, gocv.TmCcoeffNormed, gocv.NewMat())
	_, rate, _, pos := gocv.MinMaxLoc(ret)
	fmt.Println(rate)
	if rate > 0.9 {
		return []int{pos.X + (t_w / 2), pos.Y + (t_h / 2)}
	}
	return nil
}

func getDeviceId() string {
	cmd := exec.Command("adb", "devices")
	buf, _ := cmd.Output()
	out := string(buf)
	deviceList := strings.Split(out, "\n")
	deviceId := ""
	for _, v := range deviceList{
		if !strings.Contains(v, "List of devices") && strings.TrimSpace(v) != "" {
			deviceId = strings.Split(v, "\tdevice")[0]
		}
	}
	return strings.TrimSpace(deviceId)
}

func screenshot(device_id string, tag string) string {
	args := strings.Split(fmt.Sprintf("-s %+v shell screencap -p /sdcard/rhode_%+v.jpg", device_id, tag), " ")
	execute("adb", args)
	args = strings.Split(fmt.Sprintf("-s %+v pull /sdcard/rhode_%+v.jpg template/rhode_%+v.jpg", device_id, tag, tag), " ")
	execute("adb", args)
	return fmt.Sprintf("template/rhode_%+v.jpg", tag)
}

func click(device_id string, pos []int) {
	args := strings.Split(fmt.Sprintf("-s %+v shell input tap %+v %+v", pos[0], pos[1]), " ")
	execute("adb", args)
}

func lock(device_id string) {
	args := strings.Split(fmt.Sprintf("-s %+v shell input keyevent 26", device_id), " ")
	execute("adb", args)
}

func main() {
	PathExists("target", true)
	PathExists("stop", true)
	PathExists("template", true)
	device_id := getDeviceId()
	fmt.Println(fmt.Sprintf("android: %+v", device_id))
	target_list, _ := GetAllFile("target", make([]string, 0), ".jpg")
	sort.Strings(target_list)
	stop_list, _ := GetAllFile("stop", make([]string, 0), ".jpg")
	sort.Strings(stop_list)
	exit := false
	for {
		if exit {
			break
		}
		for _, tar := range target_list {
			template := screenshot(device_id, "template")
			fmt.Println(tar)
			pos := findPos(template, tar)
			if len(pos) == 2 {
				click(device_id, pos)
				time.Sleep(time.Duration(5) * time.Second)
			}
		}
		stop := screenshot(device_id, "stop")
		for _, st := range stop_list {
			ret := findPos(stop, st)
			if len(ret) == 2 {
				exit = true
				break
			}
		}
	}
	lock(device_id)
}