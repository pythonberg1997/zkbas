package transaction

import (
	"context"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMempoolTxsLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	mempool mempool.Mempool
}

func NewGetMempoolTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsLogic {
	return &GetMempoolTxsLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		mempool: mempool.New(svcCtx.Config),
	}
}
func (l *GetMempoolTxsLogic) GetMempoolTxs(req *types.ReqGetMempoolTxs) (resp *types.RespGetMempoolTxs, err error) {
	//	err := utils.CheckRequestParam(utils.TypeLimit, reflect.ValueOf(req.Limit))
	//	err = utils.CheckRequestParam(utils.TypeLimit, reflect.ValueOf(req.Limit))
	mempoolTxs, err := l.mempool.GetMempoolTxs(int64(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Error("[GetMempoolTxs] err:%v", err)
		return &types.RespGetMempoolTxs{}, err
	}

	// Todo: why not do total=len(mempoolTxs)
	total, err := l.mempool.GetMempoolTxsTotalCount()
	if err != nil {
		logx.Error("[GetMempoolTxs] err:%v", err)
		return &types.RespGetMempoolTxs{}, err
	}

	data := make([]*types.Tx, 0)
	for _, mempoolTx := range mempoolTxs {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range mempoolTx.MempoolDetails {

			if txDetail.AssetType == commonAsset.LiquidityAssetType {
				//Todo: add json string of liquidity transaction to the list
			} else {
				txDetails = append(txDetails, &types.TxDetail{
					//Todo: verify if accountBalance is still needed, since its no longer a field of table TxDetail
					//Todo: int64 or int?
					//Todo: need balance or not?  no need
					AssetId:      int(txDetail.AssetId),
					AssetType:    int(txDetail.AssetType),
					AccountIndex: int32(txDetail.AccountIndex),
					AccountName:  txDetail.AccountName,
					AccountDelta: txDetail.BalanceDelta,
				})
			}
		}
		//Todo: int64 or int?
		txAmount, err := strconv.Atoi(mempoolTx.TxAmount)
		if err != nil {
			logx.Error("[GetMempoolTxs] err:%v", err)
			return &types.RespGetMempoolTxs{}, err
		}
		// Todo: why is the field in db string?
		gasFee, err := strconv.Atoi(mempoolTx.GasFee)
		data = append(data, &types.Tx{
			TxHash:        mempoolTx.TxHash,
			TxType:        uint32(mempoolTx.TxType),
			AssetAId:      int(mempoolTx.AssetId),
			AssetBId:      int(mempoolTx.AssetId),
			TxDetails:     txDetails,
			TxAmount:      txAmount,
			NativeAddress: mempoolTx.NativeAddress,
			TxStatus:      1, //pending
			GasFeeAssetId: uint32(mempoolTx.GasFeeAssetId),
			GasFee:        int64(gasFee),
			CreatedAt:     mempoolTx.CreatedAt.Unix(),
			Memo:          mempoolTx.Memo,
		})
	}
	resp = &types.RespGetMempoolTxs{
		Total:      uint32(total),
		MempoolTxs: data,
	}
	return resp, nil
}
