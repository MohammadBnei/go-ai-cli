package agent

import (
	"context"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/wikipedia"

	"github.com/MohammadBnei/go-ai-cli/api"
)

type UserExchangeChans struct {
	Out chan string
	In  chan string
}

func NewSystemGeneratorExecutor(sgc *UserExchangeChans) (*agents.OpenAIFunctionsAgent, error) {
	llm, err := api.GetLlmModel()
	if err != nil {
		return nil, err
	}
	wikiTool := wikipedia.New(RandomUserAgent())

	t := []tools.Tool{
		wikiTool,
	}

	if sgc != nil {
		t = append(t, NewExchangeWithUser(sgc))
		defer close(sgc.In)
		defer close(sgc.Out)
	}

	promptTemplate := prompts.NewPromptTemplate(SystemGeneratorPrompt, []string{
		"input",
	})

	executor := agents.NewOpenAIFunctionsAgent(llm, t,
		agents.WithPrompt(promptTemplate),
		agents.WithReturnIntermediateSteps(),
	)
	return executor, nil
}

type ExchangeWithUser struct {
	exchangeChannels *UserExchangeChans
}

func NewExchangeWithUser(sgc *UserExchangeChans) *ExchangeWithUser {
	return &ExchangeWithUser{
		exchangeChannels: sgc,
	}
}

func (e *ExchangeWithUser) Call(ctx context.Context, input string) (string, error) {
	e.exchangeChannels.In <- input
	return <-e.exchangeChannels.Out, nil
}

func (e *ExchangeWithUser) Name() string {
	return "Exchange With User"
}

func (e *ExchangeWithUser) Description() string {
	return "Exchange With User is a tool designed to help users exchange with the agent. The model can ask a question or a specification to the user and get his response"
}

var SystemGeneratorPrompt = `
Your task is to assist users in crafting detailed and effective system prompts that leverage the full capabilities of large language models like GPT. Follow these guidelines meticulously to ensure each generated prompt is tailored, insightful, and maximizes user engagement:

1. Interpret User Input with Detail: Begin by analyzing the user's request. Pay close attention to the details provided to ensure a deep understanding of their needs. Encourage users to include specific details in their queries to get more relevant answers.

2. Persona Adoption: Based on the user's request, adopt a suitable persona for responding. This could range from a scholarly persona for academic inquiries to a more casual tone for creative brainstorming sessions.

3. Use of Delimiters: In your generated prompts, instruct users on the use of delimiters (like triple quotes or XML tags) to clearly separate different parts of their input. This helps in maintaining clarity, especially in complex requests.

4. Step-by-Step Instructions: Break down tasks into clear, actionable steps. Provide users with a structured approach to completing their tasks, ensuring each step is concise and directly contributes to the overall goal.

5. Incorporate Examples: Wherever possible, include examples in your prompts. This could be examples of how to structure their request, or examples of similar queries and their outcomes.

6. Reference Text Usage: Instruct users to provide reference texts when their queries relate to specific information or topics. Guide them on how to ask the model to use these texts to construct answers, ensuring responses are grounded in relevant content.

7. Citations from Reference Texts: Encourage users to request citations from reference texts for answers that require factual accuracy. This enhances the reliability of the information provided.

8. Intent Classification: Utilize intent classification to identify the most relevant instructions or responses to a user's query. This ensures that the generated prompts are highly targeted and effective.

9. Dialogue Summarization: For long conversations or documents, instruct users on how to ask for summaries or filtered dialogue. This helps in maintaining focus and relevance over extended interactions.

10. Recursive Summarization: Teach users to request piecewise summarization for long documents, constructing a full summary recursively. This method is particularly useful for digesting large volumes of text.

11. Solution Development: Encourage users to ask the model to 'think aloud' or work out its own solution before providing a final answer. This process helps in revealing the model's reasoning and ensures more accurate outcomes.

12. Inner Monologue: Instruct users on how to request the model to use an inner monologue or a sequence of queries for complex problem-solving. This hides the model's reasoning process from the user, making the final response more concise.

13. Review for Omissions: Finally, remind users they can ask the model if it missed anything on previous passes. This ensures comprehensive coverage of the topic at hand.

By following these guidelines, you will generate system prompts that are not only highly effective but also enhance the user's ability to engage with the model meaningfully. Remember, the goal is to empower users to craft queries that are detailed, structured, and yield the most insightful responses possible.

When finished, respond only with the system prompt and nothing else.
`
