package deviceplugin

import (
	dp "github.com/intel/cndp_device_plugin/internal/deviceplugin"
	"github.com/intel/cndp_device_plugin/internal/networking"
	"io/ioutil"
	"os"
)

/*
Fuzz sends fuzzed data into the GetConfig function
The input data is considered:
 - uninteresting if is caught by an existing error
 - interesting if it does not result in an error, input priority increases for subsequent fuzzing
 - discard if it will not unmarshall, so we don't just end up testing the json.Unmarshall function
*/
func Fuzz(data []byte) int {

	tmpfile, err := ioutil.TempFile("./", "config_")
	if err != nil {
		os.Remove(tmpfile.Name())
		panic(1) //TODO
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(data); err != nil {
		os.Remove(tmpfile.Name())
		panic(1) //TODO
	}
	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		panic(1) //TODO
	}

	_, err = dp.GetConfig(tmpfile.Name(), networking.NewHandler())
	if err != nil {
		return 0
	}

	return 1

}
