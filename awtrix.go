package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AwtrixApp struct {
	Text        string              `json:"text,omitempty"`        // The text to display
	TextCase    int                 `json:"textCase,omitempty"`    // Changes the Uppercase setting. 0=global setting, 1=forces uppercase 2=shows as it sent.
	TopText     bool                `json:"topText,omitempty"`     // Draw the text on top
	TextOffset  int                 `json:"textOffset,omitempty"`  // Sets an offset for the x position of a starting text.
	Color       string              `json:"color,omitempty"`       // The text, bar or line color
	Background  string              `json:"background,omitempty"`  // Sets a background color
	Rainbow     bool                `json:"rainbow,omitempty"`     // Fades each letter in the text differently through the entire RGB spectrum.
	Icon        string              `json:"icon,omitempty"`        // The icon ID or filename (without extension) to display on the app.
	PushIcon    int                 `json:"pushIcon,omitempty"`    // 0 = Icon doesn't move. 1 = Icon moves with text and will not appear again. 2 = Icon moves with text but appears again when the text starts to scroll again.
	Repeat      int                 `json:"repeat,omitempty"`      // Sets how many times the text should be scrolled through the matrix before the app ends.
	Duration    int                 `json:"duration,omitempty"`    // Sets how long the app or notification should be displayed.
	Hold        bool                `json:"hold,omitempty"`        // Set it to true, to hold your notification on top until you press the middle button or dismiss it via HomeAssistant. This key only belongs to notification.
	Sound       string              `json:"sound,omitempty"`       // The filename of your RTTTL ringtone file placed in the MELODIES folder (without extension).
	Rtttl       string              `json:"rtttl,omitempty"`       // Allows to send the RTTTL sound string with the json
	Bar         []int               `json:"bar,omitempty"`         // Draws a bar graph. Without icon maximum 16 values, with icon 11 values
	Line        []int               `json:"line,omitempty"`        // Draws a line chart. Without icon maximum 16 values, with icon 11 values
	Autoscale   bool                `json:"autoscale,omitempty"`   // Enables or disables autoscaling for bar and line chart
	Progress    int                 `json:"progress,omitempty"`    // Shows a progressbar. Value can be 0-100
	ProgressC   string              `json:"progressC,omitempty"`   // The color of the progressbar
	ProgressBC  string              `json:"progressBC,omitempty"`  // The color of the progressbar background
	Pos         int                 `json:"pos,omitempty"`         // Defines the position of your custom page in the loop, starting at 0 for the first position. This will only apply with your first push. This function is experimental
	Draw        []AwtrixDrawCommand `json:"draw,omitempty"`        // Array of drawing instructions. Each object represents a drawing command.
	Lifetime    int                 `json:"lifetime,omitempty"`    // Removes the custom app when there is no update after the given time in seconds
	Stack       bool                `json:"stack,omitempty"`       // Defines if the notification will be stacked. false will immediately replace the current notification
	Wakeup      bool                `json:"wakeup,omitempty"`      // If the Matrix is off, the notification will wake it up for the time of the notification.
	NoScroll    bool                `json:"noScroll,omitempty"`    // Disables the text scrolling
	Clients     []string            `json:"clients,omitempty"`     // Allows to forward a notification to other awtrix. Use the MQTT prefix for MQTT and IP addresses for HTTP
	ScrollSpeed int                 `json:"scrollSpeed,omitempty"` // Modifies the scroll speed. You need to enter a percentage value
}

type AwtrixDrawCommand struct {
	Command string
	X       int
	Y       int
	Width   int
	Height  int
	Text    string
	Color   string
}

func (u *AwtrixDrawCommand) MarshalJSON() ([]byte, error) {
	switch u.Command {
	case "dt":
		return marshalDrawText(*u), nil
	case "df":
		return marshalDrawRectangleFilled(*u), nil
	default:
		return []byte{}, fmt.Errorf("unknown draw command: %s", u.Command)
	}
}

func marshalDrawText(command AwtrixDrawCommand) []byte {
	str := fmt.Sprintf("{\"dt\": [%d, %d, \"%s\", \"%s\"]}", command.X, command.Y, command.Text, command.Color)
	return []byte(str)
}

func marshalDrawRectangleFilled(command AwtrixDrawCommand) []byte {
	str := fmt.Sprintf("{\"df\": [%d, %d, %d, %d, \"%s\"]}", command.X, command.Y, command.Width, command.Height, command.Color)
	return []byte(str)
}

func postNotification(ip string, app AwtrixApp) error {
	postURL, err := url.ParseRequestURI(fmt.Sprintf("http://%s/api/notify", ip))
	if err != nil {
		return err
	}

	body, err := json.Marshal(app)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", postURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	return doRequest(request)
}

func postApplication(ip string, name string, app AwtrixApp) error {
	postURL, err := url.ParseRequestURI(fmt.Sprintf("http://%s/api/custom", ip))
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("name", name)
	postURL.RawQuery = params.Encode()

	body, err := json.Marshal(app)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", postURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	return doRequest(request)
}

func doRequest(request *http.Request) error {
	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if string(resBody) != "OK" {
		return errors.New(string(resBody))
	}
	return nil
}
