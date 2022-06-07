package main

import (
	"context"
	"flag"
	zecreyLegend "github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey-legend"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/sender/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/sender/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/sender/internal/svc"
	"math/big"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f",
	"D:\\Projects\\mygo\\src\\Zecrey\\SherLzp\\zecrey-legend\\service\\cronjob\\sender\\etc\\local.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)
	// srv := server.NewSenderServer(ctx)

	networkEndpointName := c.ChainConfig.NetworkRPCSysConfigName
	networkEndpoint, err := ctx.SysConfigModel.GetSysconfigByName(networkEndpointName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch networkEndpoint from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	ZecreyRollupAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.ZecreyContractAddrSysConfigName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch ZecreyRollupAddress from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.ZecreyContractAddrSysConfigName)
		panic(err)
	}

	cli, err := _rpc.NewClient(networkEndpoint.Value)
	if err != nil {
		panic(err)
	}
	var chainId *big.Int
	if c.ChainConfig.L1ChainId == "" {
		chainId, err = cli.ChainID(context.Background())
		if err != nil {
			panic(err)
		}
	} else {
		var (
			isValid bool
		)
		chainId, isValid = new(big.Int).SetString(c.ChainConfig.L1ChainId, 10)
		if !isValid {
			panic("invalid l1 chain id")
		}
	}

	authCli, err := _rpc.NewAuthClient(cli, c.ChainConfig.Sk, chainId)
	if err != nil {
		panic(err)
	}
	zecreyInstance, err := zecreyLegend.LoadZecreyLegendInstance(cli, ZecreyRollupAddress.Value)
	if err != nil {
		panic(err)
	}
	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}

	var param = &logic.SenderParam{
		Cli:                  cli,
		AuthCli:              authCli,
		ZecreyLegendInstance: zecreyInstance,
		MaxWaitingTime:       c.ChainConfig.MaxWaitingTime * time.Second.Milliseconds(),
		MaxBlocksCount:       3,
		GasPrice:             gasPrice,
		GasLimit:             c.ChainConfig.GasLimit,
	}

	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {

		logx.Info("========================= start sender committer task =========================")
		err := logic.SendCommittedBlocks(
			param,
			ctx.L1TxSenderModel,
			ctx.BlockModel,
			ctx.BlockForCommitModel,
		)
		if err != nil {
			logx.Info("[sender.SendCommittedBlocks main] unable to run:", err)
		}
		logx.Info("========================= end sender committer task =========================")

		logx.Info("========================= start sender verifier task =========================")
		//err = logic.SendVerifiedAndExecutedBlocks(param, ctx.L1TxSenderModel, ctx.ProofSenderModel)
		//if err != nil {
		//	logx.Info("[sender.SendCommittedBlocks main] unable to run:", err)
		//}
		logx.Info("========================= end sender verifier task =========================")

	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("sender cron job is starting......")
	select {}
}