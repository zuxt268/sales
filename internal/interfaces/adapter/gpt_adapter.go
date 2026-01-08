package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/model"
)

type GptAdapter interface {
	Analyze(ctx context.Context, domain *model.Domain) error
	AnalyzeSiteIndustry(ctx context.Context, text string) (string, error)
}

type gptAdapter struct {
	client *openai.Client
}

func NewGptAdapter() GptAdapter {
	apiKey := config.Env.OpenaiApiKey
	client := openai.NewClient(apiKey)
	return &gptAdapter{
		client: client,
	}
}

const systemPrompt = `
業種リストです。この中から業種を選らんでください。
農業
林業
漁業（水産養殖業を除く）
水産養殖業
鉱業，採石業，砂利採取業
総合工事業
職別工事業(設備工事業を除く)
設備工事業
食料品製造業
飲料・たばこ・飼料製造業
繊維工業
木材・木製品製造業（家具を除く）
家具・装備品製造業
パルプ・紙・紙加工品製造業
印刷・同関連業
化学工業
石油製品・石炭製品製造業
プラスチック製品製造業（別掲を除く）
ゴム製品製造業
なめし革・同製品・毛皮製造業
窯業・土石製品製造業
鉄鋼業
非鉄金属製造業
金属製品製造業
はん用機械器具製造業
生産用機械器具製造業
業務用機械器具製造業
電子部品・デバイス・電子回路製造業
電気機械器具製造業
情報通信機械器具製造業
輸送用機械器具製造業
その他の製造業
電気業
ガス業
熱供給業
水道業
通信業
放送業
情報サービス業
インターネット附随サービス業
映像・音声・文字情報制作業
鉄道業
道路旅客運送業
道路貨物運送業
水運業
航空運輸業
倉庫業
運輸に附帯するサービス業
郵便業（信書便事業を含む）
各種商品卸売業
繊維・衣服等卸売業
飲食料品卸売業
建築材料，鉱物・金属材料等卸売業
機械器具卸売業
その他の卸売業
各種商品小売業
織物・衣服・身の回り品小売業
飲食料品小売業
機械器具小売業
その他の小売業
無店舗小売業
銀行業
協同組織金融業
貸金業，クレジットカード業等非預金信用機関
金融商品取引業，商品先物取引業
補助的金融業等
保険業（保険媒介代理業，保険サービス業を含む）
不動産取引業
不動産賃貸業・管理業
物品賃貸業
学術・開発研究機関
専門サービス業（他に分類されないもの）
広告業
技術サービス業（他に分類されないもの）
宿泊業
飲食店
持ち帰り・配達飲食サービス業
洗濯・理容・美容・浴場業
その他の生活関連サービス業
娯楽業
学校教育
その他の教育，学習支援業
医療業
保健衛生
社会保険・社会福祉・介護事業
郵便局
協同組合（他に分類されないもの）
廃棄物処理業
自動車整備業
機械等修理業（別掲を除く）
職業紹介・労働者派遣業
その他の事業サービス業
政治・経済・文化団体
宗教
その他のサービス業
外国公務
国家公務
地方公務
分類不能の産業
`

const promptTemplate = `"""%s"""
以上の情報から、業種、代表者名、会社名、都道府県を答えてください。単語で答えてください。見つからない場合は、「なし」と表示してください。
４つをカンマ区切りで一行で出力してください。
順番は守ってください。
例）自動車整備業,山田太郎,株式会社タロウ,東京都
`

func (a *gptAdapter) Analyze(ctx context.Context, d *model.Domain) error {
	prompt := fmt.Sprintf(promptTemplate, d.RawPage)
	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-5-nano",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("ChatCompletion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil
	}
	contents := strings.Split(resp.Choices[0].Message.Content, ",")
	if len(contents) != 4 {
		return nil
	}
	fmt.Println(contents)
	if contents[0] != "なし" {
		d.Industry = contents[0]
	}
	if contents[1] != "なし" {
		d.President = contents[1]
	}
	if contents[2] != "なし" {
		d.Company = contents[2]
	}
	if contents[3] != "なし" {
		d.Prefecture = contents[3]
	}
	return nil
}

const promptAnalyzeIndustryTemplate = `"""%s"""
以上の情報から、業種を答えてください。単語で答えてください。一番近いものを選んでください。
できるだけ具体的なものを選んでください。できれば1個で答えてください。
1個に定められない場合は、3個まで答えてください。その場合、カンマ区切りで答えてください。
例）自動車整備業,広告業
`

func (a *gptAdapter) AnalyzeSiteIndustry(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf(promptAnalyzeIndustryTemplate, text)
	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-5-nano",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}
