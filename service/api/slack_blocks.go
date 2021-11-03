package api

// SlackBlocks are defined in the documentation here: https://api.slack.com/messaging/composing/layouts#add_blocks_array
type SlackBlocks struct {
	Blocks interface{} `json:"blocks"`
}

type SlackBlock struct {
	BlockId   string               `json:"block_id,omitempty"`
	Type      string               `json:"type"`
	Text      interface{}          `json:"text,omitempty"`
	Accessory *SlackBlockAccessory `json:"accessory,omitempty"`
	Element   interface{}          `json:"element,omitempty"`
	Label     *SlackBlockLabel     `json:"label,omitempty"`
}

type SlackBlockLabel struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji"`
}

type SlackFieldsBlock struct {
	Type   string        `json:"type"`
	Fields []interface{} `json:"fields"`
}

type SlackBlockRawType struct {
	Type string `json:"type"`
}

type SlackBlockAccessory struct {
	ActionId    string      `json:"action_id"`
	Type        string      `json:"type"`
	Placeholder interface{} `json:"placeholder,omitempty"`
	Text        interface{} `json:"text,omitempty"`
}

type SlackBlockSubmit struct {
	Type string      `json:"type"`
	Text interface{} `json:"text,omitempty"`
}

// Create this i think: https://api.slack.com/reference/block-kit/block-elements#users_select__example
// [
//   {
//     "type": "section",
//     "block_id": "section678",
//     "text": {
//       "type": "mrkdwn",
//       "text": "Pick a user from the dropdown list"
//     },
//     "accessory": {
//       "action_id": "text1234",
//       "type": "users_select",
//       "placeholder": {
//         "type": "plain_text",
//         "text": "Select an item"
//       }
//     }
//   }
// ]
