package miraitool

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	hertzc "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
)

func UpGroupImgByUrl(c *client.QQClient, groupCode int64, url string) (*message.GroupImageElement, error) {
	_, cc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cc()
	req := protocol.Request{}
	res := protocol.Response{}
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.192 Safari/537.36")
	req.SetRequestURI(url)
	err := hertzc.DoTimeout(context.Background(), &req, &res, 10*time.Second)
	if err != nil {
		return nil, err
	}
	img, err := c.UploadGroupImage(groupCode, bytes.NewReader(res.BodyBytes()))
	if err != nil {
		log.Println("UpGroupImgByUrl Err:", err)
		return nil, nil
	}
	return img, nil
}

func UpGroupFile(c *client.QQClient, groupCode int64, filename string) (*message.ShortVideoElement, error) {
	f, _ := os.Open(filename)
	defer func() {
		os.Remove(filename)
		os.Remove(filename + ".jpg")
	}()
	defer f.Close()

	data, _ := os.ReadFile(filename + ".jpg")
	thumb := bytes.NewReader(data)
	// _, _ = f.Seek(0, io.SeekStart)
	// _, _ = thumb.Seek(0, io.SeekStart)

	img, err := c.UploadGroupShortVideo(groupCode, f, thumb)
	if err != nil {
		log.Println("upFile Err:", err)
		return nil, nil
	}
	return img, nil
}

// ExtractCover 获取给定视频文件的Cover
func extractCover(src string, target string) error {
	// cmd := exec.Command("ffmpeg", "-i", src, "-y", "-r", "1", "-f", "image2", target)
	cmd := exec.Command("ffmpeg", "-i", src, "-y", "-f", "image2", "-frames", "1", target)
	return errors.Wrap(cmd.Run(), "extract video cover failed")
}

func YouGet(url string) (name string) {
	cmd1 := exec.Command("you-get", url)
	cmd1.Run()
	out := bytes.NewBuffer(nil)
	cmd2 := exec.Command("you-get", "--json", url)
	cmd2.Stdout = out
	cmd2.Run()

	fmt.Println(out.String())

	name = gjson.GetBytes(out.Bytes(), "title").Str + ".mp4"

	cmd := exec.Command("ffmpeg", "-i", name, "-y", "-f", "image2", "-frames", "1", name+".jpg")
	cmd.Run()
	return
}
