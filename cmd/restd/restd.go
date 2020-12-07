package main
import (
	"os/user"
	"strconv"

	"github.com/untangle/restd/services/gind"
)

func main() {
	userinfo, err := user.Current()
	if err != nil {
		panic(err)
	}

	userid, err := strconv.Atoi(userinfo.Uid)
	if err != nil {
		panic(err)
	}

	if userid != 0 {
		panic("This application must be run as root")
	}

	gind.Startup()

}
