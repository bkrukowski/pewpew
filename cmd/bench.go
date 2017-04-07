package cmd

import (
	"errors"
	"fmt"

	pewpew "github.com/bengadbois/pewpew/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var benchCmd = &cobra.Command{
	Use:   "bench URL...",
	Short: "Run benchmark tests",
	RunE: func(cmd *cobra.Command, args []string) error {

		benchCfg := pewpew.BenchConfig{}
		err := viper.Unmarshal(&benchCfg)
		if err != nil {
			fmt.Println(err)
			return errors.New("could not parse config file")
		}

		//global configs
		benchCfg.Quiet = viper.GetBool("quiet")
		benchCfg.Verbose = viper.GetBool("verbose")
		benchCfg.BeginRPS = viper.GetInt("begin")
		benchCfg.EndRPS = viper.GetInt("end")
		benchCfg.Interval = viper.GetInt("interval")
		benchCfg.Duration = viper.GetString("duration")
		benchCfg.Cooldown = viper.GetString("cooldown")

		//URLs are handled differently that other config options
		//command line specifying URLs take higher precedence than config URLs

		//check either set via config or command line
		if len(benchCfg.BenchTargets) == 0 && len(args) < 1 {
			return errors.New("requires URL")
		}

		//if URLs are set on command line, use that for Targets instead of config
		if len(args) >= 1 {
			benchCfg.BenchTargets = make([]pewpew.BenchTarget, len(args))
			for i := range benchCfg.BenchTargets {
				benchCfg.BenchTargets[i].Target.URL = args[i]
				//use global configs instead of the config file's individual target settings
				benchCfg.BenchTargets[i].Target.RegexURL = viper.GetBool("regex")
				benchCfg.BenchTargets[i].Target.Timeout = viper.GetString("timeout")
				benchCfg.BenchTargets[i].Target.Method = viper.GetString("request-method")
				benchCfg.BenchTargets[i].Target.Body = viper.GetString("body")
				benchCfg.BenchTargets[i].Target.BodyFilename = viper.GetString("body-file")
				benchCfg.BenchTargets[i].Target.Headers = viper.GetString("headers")
				benchCfg.BenchTargets[i].Target.Cookies = viper.GetString("cookies")
				benchCfg.BenchTargets[i].Target.UserAgent = viper.GetString("user-agent")
				benchCfg.BenchTargets[i].Target.BasicAuth = viper.GetString("basic-auth")
				benchCfg.BenchTargets[i].Target.Compress = viper.GetBool("compress")
				benchCfg.BenchTargets[i].Target.KeepAlive = viper.GetBool("keepalive")
				benchCfg.BenchTargets[i].Target.FollowRedirects = viper.GetBool("follow-redirects")
				benchCfg.BenchTargets[i].Target.NoHTTP2 = viper.GetBool("no-http2")
				benchCfg.BenchTargets[i].Target.EnforceSSL = viper.GetBool("enforce-ssl")
			}
		} else {
			//set non-URL target settings
			//walk through viper.Get() because that will show which were
			//explictly set instead of guessing at zero-valued defaults
			for i, target := range viper.Get("targets").([]interface{}) {
				targetMapVals := target.(map[string]interface{})
				if _, set := targetMapVals["RegexURL"]; !set {
					benchCfg.BenchTargets[i].Target.RegexURL = viper.GetBool("regex")
				}
				if _, set := targetMapVals["Timeout"]; !set {
					benchCfg.BenchTargets[i].Target.Timeout = viper.GetString("timeout")
				}
				if _, set := targetMapVals["Method"]; !set {
					benchCfg.BenchTargets[i].Target.Method = viper.GetString("method")
				}
				if _, set := targetMapVals["Body"]; !set {
					benchCfg.BenchTargets[i].Target.Body = viper.GetString("body")
				}
				if _, set := targetMapVals["BodyFilename"]; !set {
					benchCfg.BenchTargets[i].Target.BodyFilename = viper.GetString("bodyFile")
				}
				if _, set := targetMapVals["Headers"]; !set {
					benchCfg.BenchTargets[i].Target.Headers = viper.GetString("headers")
				}
				if _, set := targetMapVals["Cookies"]; !set {
					benchCfg.BenchTargets[i].Target.Cookies = viper.GetString("cookies")
				}
				if _, set := targetMapVals["UserAgent"]; !set {
					benchCfg.BenchTargets[i].Target.UserAgent = viper.GetString("userAgent")
				}
				if _, set := targetMapVals["BasicAuth"]; !set {
					benchCfg.BenchTargets[i].Target.BasicAuth = viper.GetString("basicAuth")
				}
				if _, set := targetMapVals["Compress"]; !set {
					benchCfg.BenchTargets[i].Target.Compress = viper.GetBool("compress")
				}
				if _, set := targetMapVals["KeepAlive"]; !set {
					benchCfg.BenchTargets[i].Target.KeepAlive = viper.GetBool("keepalive")
				}
				if _, set := targetMapVals["FollowRedirects"]; !set {
					benchCfg.BenchTargets[i].Target.FollowRedirects = viper.GetBool("followredirects")
				}
				if _, set := targetMapVals["NoHTTP2"]; !set {
					benchCfg.BenchTargets[i].Target.NoHTTP2 = viper.GetBool("no-http2")
				}
				if _, set := targetMapVals["EnforceSSL"]; !set {
					benchCfg.BenchTargets[i].Target.EnforceSSL = viper.GetBool("enforce-ssl")
				}
			}
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(benchCmd)
	benchCmd.Flags().IntP("begin", "b", pewpew.DefaultBeginRPS, "Beginning requests per second.")
	viper.BindPFlag("begin", benchCmd.Flags().Lookup("begin"))

	benchCmd.Flags().IntP("end", "e", pewpew.DefaultEndRPS, "Ending requests per second.")
	viper.BindPFlag("end", benchCmd.Flags().Lookup("end"))

	benchCmd.Flags().IntP("interval", "i", pewpew.DefaultInterval, "How many requests per second to add per round until the end.")
	viper.BindPFlag("interval", benchCmd.Flags().Lookup("interval"))

	benchCmd.Flags().StringP("duration", "d", pewpew.DefaultDuration, "How long each round lasts.")
	viper.BindPFlag("duration", benchCmd.Flags().Lookup("duration"))

	benchCmd.Flags().String("cooldown", pewpew.DefaultCooldown, "How long to pause between rounds.")
	viper.BindPFlag("cooldown", benchCmd.Flags().Lookup("cooldown"))
}
