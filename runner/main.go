package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ChristopherHX/github-act-runner/actionsrunner"
	"github.com/ChristopherHX/github-act-runner/protocol"
	"github.com/ChristopherHX/github-act-runner/runnerconfiguration/compat"
	"github.com/actions-oss/act-cli/pkg/exprparser"
)

type NoOpt struct {
}

// GetInput implements [runnerconfiguration.Survey].
func (n NoOpt) GetInput(prompt string, def string) string {
	return ""
}

// GetMultiSelectInput implements [runnerconfiguration.Survey].
func (n NoOpt) GetMultiSelectInput(prompt string, options []string) []string {
	return nil
}

// GetSelectInput implements [runnerconfiguration.Survey].
func (n NoOpt) GetSelectInput(prompt string, options []string, def string) string {
	return ""
}

type RunnerEnv struct {
}

// ExecWorker implements [actionsrunner.RunnerEnvironment].
func (r RunnerEnv) ExecWorker(run *actionsrunner.RunRunner, wc actionsrunner.WorkerContext, jobreq *protocol.AgentJobRequestMessage, src []byte) error {
	wc.Logger().Log(fmt.Sprintf("%vExpect jobid", time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z ")))
	wc.Logger().Update()
	wc.Logger().Log(fmt.Sprintf("%v##[Error]No matching webhook received within 5 Minutes, killing runner", time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z ")))
	eval := &exprparser.EvaluationEnvironment{
		CtxData: map[string]interface{}{
			"ci": "hello world",
		},
	}
	intp := exprparser.NewInterpeter(eval, exprparser.Config{})
	res, _ := intp.Evaluate("tojson(tojson(ci))", exprparser.DefaultStatusCheckNone)
	wc.Logger().Log(fmt.Sprintf("%v%s", time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z "), res))
	// wc.Message().
	for i, step := range wc.Message().Steps {
		timelineEntry := protocol.CreateTimelineEntry(jobreq.JobID, step.ID, "Run "+step.Reference.Name+"@"+step.Reference.Ref)
		stepEntry := wc.Logger().Append(&timelineEntry)
		stepEntry.Order = int32(i)
	}
	wc.Logger().Current().Complete("Succeeded")
	wc.Logger().Update()
	for _, step := range wc.Message().Steps {
		wc.Logger().MoveNextExt(true)
		inputs := step.Inputs.ToJSONRawObject()
		if m, ok := inputs.(map[string]interface{}); ok {
			if sURL, ok := m["url"].(string); ok {
				wc.Logger().Log(fmt.Sprintf("%vExecute Http Request %v", time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z "), sURL))

				req, err := http.NewRequestWithContext(wc.JobExecCtx(), http.MethodGet, sURL, nil)
				if err != nil {
					println(err)
					wc.Logger().Current().Complete("Failure")
				}
				req.SetBasicAuth("", wc.Message().Variables["system.github.token"].Value)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					println(err)
					wc.Logger().Current().Complete("Failure")
				}
				data, _ := io.ReadAll(resp.Body)
				wc.Logger().Log(fmt.Sprintf("%vStatus %v, Response: %v", time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z "), resp.StatusCode, string(data)))
				wc.Logger().Current().Complete("Succeeded")
				continue
			}
		}
		wc.Logger().Current().Complete("Failure")
	}
	wc.Logger().MoveNextExt(false)
	wc.Logger().TimelineRecords.Value[0].Complete("Succeeded")
	wc.Logger().Finish()
	wc.FinishJob("Succeeded", &map[string]protocol.VariableValue{})
	return nil
}

// Printf implements [actionsrunner.RunnerEnvironment].
func (r RunnerEnv) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// ReadJSON implements [actionsrunner.RunnerEnvironment].
func (r RunnerEnv) ReadJSON(fname string, obj interface{}) error {
	return nil
}

// Remove implements [actionsrunner.RunnerEnvironment].
func (r RunnerEnv) Remove(fname string) error {
	return nil
}

// WriteJSON implements [actionsrunner.RunnerEnvironment].
func (r RunnerEnv) WriteJSON(fname string, obj interface{}) error {
	return nil
}

func main() {
	// println("Start 2")
	// <-time.After(time.Second)
	// println("Done 2")
	// return
	cl := http.Client{}
	// configRunner := &runnerconfiguration.ConfigureRunner{
	// 	ConfigureRemoveRunner: runnerconfiguration.ConfigureRemoveRunner{
	// 		Client:     &cl,
	// 		URL:        "https://github.com/ChristopherHX/github-act-runner-test",
	// 		Token:      "hhhhhhhhhhhhhh",
	// 		Unattended: true,
	// 		Trace:      true,
	// 	},
	// 	Labels:          []string{"wasm"},
	// 	NoDefaultLabels: true,
	// 	Ephemeral:       true,
	// }
	// auth, _ := configRunner.Authenticate(&cl, NoOpt{})
	// res, _ := configRunner.Configure(&runnerconfiguration.RunnerSettings{}, NoOpt{}, auth)
	// _ = res
	// res := &instance.RunnerSettings{}
	// data, _ := json.MarshalIndent(res, "", "  ")
	// println(string(data))
	res, _ := compat.ParseJitRunnerConfig(os.Getenv("JIT_CONFIG"))
	_ = cl

	runner := &actionsrunner.RunRunner{
		Once:     true,
		Settings: res,
		Trace:    true,
		Version:  "3.0.0",
	}
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	// defer cancel()
	ctx := context.Background()
	runner.Run(RunnerEnv{}, ctx, ctx)
}
