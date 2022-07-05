package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	defaultSwapFee    = sdk.MustNewDecFromStr("0.025")
	defaultExitFee    = sdk.MustNewDecFromStr("0.025")
	defaultPoolParams = balancer.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultFutureGovernor = ""

	// pool assets
	defaultFooAsset = balancertypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}
	defaultBarAsset = balancertypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}
	defaultPoolAssets           = []balancertypes.PoolAsset{defaultFooAsset, defaultBarAsset}
	defaultAcctFunds  sdk.Coins = sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000)),
	)
)

func (suite *KeeperTestSuite) TestCreateBalancerPool() {
	params := suite.App.GAMMKeeper.GetParams(suite.Ctx)

	// get raw pool creation fee(s) as DecCoins
	poolCreationFeeDecCoins := sdk.DecCoins{}
	for _, coin := range params.PoolCreationFee {
		poolCreationFeeDecCoins = poolCreationFeeDecCoins.Add(sdk.NewDecCoin(coin.Denom, coin.Amount))
	}

	// we keep this account empty to test attempts with insufficient balance
	emptySender := suite.TestAccs[1]

	// TODO: Refactor this to be more sensible.
	// The struct should contain a MsgCreateBalancerPool.
	// then the scaffolding should test pool creation, and check if it was a success or not.
	// (And should be moved to balancer package)
	// PoolCreationFee tests should get their own isolated test in this package.
	tests := []struct {
		name       string
		msg        balancertypes.MsgCreateBalancerPool
		expectPass bool
	}{
		{
			name:       "create pool with default assets",
			msg:        balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], defaultPoolParams, defaultPoolAssets, defaultFutureGovernor),
			expectPass: true,
		}, {
			name:       "create pool with no assets",
			msg:        balancer.NewMsgCreateBalancerPool(emptySender, defaultPoolParams, defaultPoolAssets, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create a pool with negative swap fee",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(-1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, defaultPoolAssets, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create a pool with negative exit fee",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(-1, 2),
			}, defaultPoolAssets, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create the pool with empty PoolAssets",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{}, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create the pool with 0 weighted PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(0),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create the pool with negative weighted PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(-1),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create the pool with 0 balance PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(0)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create the pool with negative balance PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token: sdk.Coin{
					Denom:  "foo",
					Amount: sdk.NewInt(-1),
				},
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			expectPass: false,
		}, {
			name: "create the pool with duplicated PoolAssets",
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		keeper := suite.App.GAMMKeeper

		// fund sender test account
		sender, err := sdk.AccAddressFromBech32(test.msg.Sender)
		suite.Require().NoError(err, "test: %v", test.name)
		if sender.String() != emptySender.String() {
			suite.FundAcc(sender, defaultAcctFunds)
		}

		// note starting balances for community fee pool and pool creator account
		feePoolBeforeNewPool := suite.App.DistrKeeper.GetFeePoolCommunityCoins(suite.Ctx)
		senderBalBeforeNewPool := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sender)

		// attempt to create a pool with the given NewMsgCreateBalancerPool message
		poolId, err := keeper.CreatePool(suite.Ctx, test.msg)

		if test.expectPass {
			suite.Require().NoError(err, "test: %v", test.name)

			// check to make sure new pool exists and has minted the correct number of pool shares
			pool, err := keeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
				fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply.String()),
			)

			// make sure pool creation fee is correctly sent to community pool
			feePool := suite.App.DistrKeeper.GetFeePoolCommunityCoins(suite.Ctx)
			suite.Require().Equal(feePool, feePoolBeforeNewPool.Add(poolCreationFeeDecCoins...))

			// get expected tokens in new pool and corresponding pool shares
			expectedPoolTokens := sdk.Coins{}
			for _, asset := range test.msg.GetPoolAssets() {
				expectedPoolTokens = expectedPoolTokens.Add(asset.Token)
			}
			expectedPoolShares := sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)

			// make sure sender's balance is updated correctly
			senderBal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sender)
			expectedSenderBal := senderBalBeforeNewPool.Sub(params.PoolCreationFee).Sub(expectedPoolTokens).Add(expectedPoolShares)
			suite.Require().Equal(senderBal.String(), expectedSenderBal.String())

			// check pool's liquidity is correctly increased
			liquidity := suite.App.GAMMKeeper.GetTotalLiquidity(suite.Ctx)
			suite.Require().Equal(expectedPoolTokens.String(), liquidity.String())
		} else {
			suite.Require().Error(err, "test: %v", test.name)
		}
	}
}

func (suite *KeeperTestSuite) TestPoolCreationFee() {
	params := suite.App.GAMMKeeper.GetParams(suite.Ctx)

	// get raw pool creation fee(s) as DecCoins
	poolCreationFeeDecCoins := sdk.DecCoins{}
	for _, coin := range params.PoolCreationFee {
		poolCreationFeeDecCoins = poolCreationFeeDecCoins.Add(sdk.NewDecCoin(coin.Denom, coin.Amount))
	}

	tests := []struct {
		name            string
		poolCreationFee sdk.Coins
		msg             balancertypes.MsgCreateBalancerPool
		expectPass      bool
	}{
		{
			name:            "no pool creation fee for default asset pool",
			poolCreationFee: sdk.Coins{},
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, defaultPoolAssets, defaultFutureGovernor),
			expectPass: true,
		}, {
			name:            "nil pool creation fee on basic pool",
			poolCreationFee: nil,
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, defaultPoolAssets, defaultFutureGovernor),
			expectPass: true,
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		keeper := suite.App.GAMMKeeper

		// set pool creation fee
		keeper.SetParams(suite.Ctx, types.Params{
			PoolCreationFee: test.poolCreationFee,
		})

		// fund sender test account
		sender, err := sdk.AccAddressFromBech32(test.msg.Sender)
		suite.Require().NoError(err, "test: %v", test.name)
		suite.FundAcc(sender, defaultAcctFunds)

		// note starting balances for community fee pool and pool creator account
		feePoolBeforeNewPool := suite.App.DistrKeeper.GetFeePoolCommunityCoins(suite.Ctx)
		senderBalBeforeNewPool := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sender)

		// attempt to create a pool with the given NewMsgCreateBalancerPool message
		poolId, err := keeper.CreatePool(suite.Ctx, test.msg)

		if test.expectPass {
			suite.Require().NoError(err, "test: %v", test.name)

			// check to make sure new pool exists and has minted the correct number of pool shares
			pool, err := keeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
				fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply.String()),
			)

			// make sure pool creation fee is correctly sent to community pool
			feePool := suite.App.DistrKeeper.GetFeePoolCommunityCoins(suite.Ctx)
			suite.Require().Equal(feePool, feePoolBeforeNewPool.Add(sdk.NewDecCoinsFromCoins(test.poolCreationFee...)...))
			// get expected tokens in new pool and corresponding pool shares
			expectedPoolTokens := sdk.Coins{}
			for _, asset := range test.msg.GetPoolAssets() {
				expectedPoolTokens = expectedPoolTokens.Add(asset.Token)
			}
			expectedPoolShares := sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)

			// make sure sender's balance is updated correctly
			senderBal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, sender)
			expectedSenderBal := senderBalBeforeNewPool.Sub(test.poolCreationFee).Sub(expectedPoolTokens).Add(expectedPoolShares)
			suite.Require().Equal(senderBal.String(), expectedSenderBal.String())

			// check pool's liquidity is correctly increased
			liquidity := suite.App.GAMMKeeper.GetTotalLiquidity(suite.Ctx)
			suite.Require().Equal(expectedPoolTokens.String(), liquidity.String())
		} else {
			suite.Require().Error(err, "test: %v", test.name)
		}
	}
}

// TODO: Add more edge cases around TokenInMaxs not containing every token in pool.
func (suite *KeeperTestSuite) TestJoinPoolNoSwap() {
	tests := []struct {
		fn func(poolId uint64)
	}{
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper
				balancesBefore := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.TestAccs[1])
				err := keeper.JoinPoolNoSwap(suite.Ctx, suite.TestAccs[1], poolId, types.OneShare.MulRaw(50), sdk.Coins{})
				suite.Require().NoError(err)
				suite.Require().Equal(types.OneShare.MulRaw(50).String(), suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[1], "gamm/pool/1").Amount.String())
				balancesAfter := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.TestAccs[1])

				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100000000gamm/pool/1.
				// Thus, to get the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be provided.
				suite.Require().Equal("5000", deltaBalances.AmountOf("foo").String())
				suite.Require().Equal("5000", deltaBalances.AmountOf("bar").String())

				liquidity := suite.App.GAMMKeeper.GetTotalLiquidity(suite.Ctx)
				suite.Require().Equal("15000bar,15000foo", liquidity.String())
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper
				err := keeper.JoinPoolNoSwap(suite.Ctx, suite.TestAccs[1], poolId, sdk.NewInt(0), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting 0 share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper
				err := keeper.JoinPoolNoSwap(suite.Ctx, suite.TestAccs[1], poolId, sdk.NewInt(-1), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting negative share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper
				// Test the "tokenInMaxs"
				// In this case, to get the 50 * OneShare amount of share token, the foo, bar token are expected to be provided as 5000 amounts.
				err := keeper.JoinPoolNoSwap(suite.Ctx, suite.TestAccs[1], poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("bar", sdk.NewInt(4999)), sdk.NewCoin("foo", sdk.NewInt(4999)),
				})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper
				// Test the "tokenInMaxs"
				// In this case, to get the 50 * OneShare amount of share token, the foo, bar token are expected to be provided as 5000 amounts.
				err := keeper.JoinPoolNoSwap(suite.Ctx, suite.TestAccs[1], poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("bar", sdk.NewInt(5000)), sdk.NewCoin("foo", sdk.NewInt(5000)),
				})
				suite.Require().NoError(err)

				liquidity := suite.App.GAMMKeeper.GetTotalLiquidity(suite.Ctx)
				suite.Require().Equal("15000bar,15000foo", liquidity.String())
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper
				// Test the "tokenInMaxs" with an additional invalid denom
				// In this case, to get the 50 * OneShare amount of share token, the foo, bar token are expected to be provided as 5000 amounts.
				// The test input has the correct amount for each, but also includes an incorrect denom that should cause an error
				err := keeper.JoinPoolNoSwap(suite.Ctx, suite.TestAccs[1], poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("bar", sdk.NewInt(5000)), sdk.NewCoin("foo", sdk.NewInt(5000)), sdk.NewCoin("baz", sdk.NewInt(5000)),
				})
				suite.Require().Error(err)
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range suite.TestAccs {
			suite.FundAcc(acc, defaultAcctFunds)
		}

		// Create the pool at first
		msg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, defaultPoolAssets, defaultFutureGovernor)
		poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
		suite.Require().NoError(err)

		test.fn(poolId)
	}
}

func (suite *KeeperTestSuite) TestExitPool() {
	tests := []struct {
		fn func(poolId uint64)
	}{
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper
				// Acc2 has no share token.
				_, err := keeper.ExitPool(suite.Ctx, suite.TestAccs[1], poolId, types.OneShare.MulRaw(50), sdk.Coins{})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper

				balancesBefore := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.TestAccs[0])
				_, err := keeper.ExitPool(suite.Ctx, suite.TestAccs[0], poolId, types.InitPoolSharesSupply.QuoRaw(2), sdk.Coins{})
				suite.Require().NoError(err)
				// (100 - 50) * OneShare should remain.
				suite.Require().Equal(types.InitPoolSharesSupply.QuoRaw(2).String(), suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], "gamm/pool/1").Amount.String())
				balancesAfter := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.TestAccs[0])

				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100*OneShare gamm/pool/1.
				// Thus, to refund the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be refunded.
				suite.Require().Equal("-5000", deltaBalances.AmountOf("foo").String())
				suite.Require().Equal("-5000", deltaBalances.AmountOf("bar").String())
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper

				_, err := keeper.ExitPool(suite.Ctx, suite.TestAccs[0], poolId, sdk.NewInt(0), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting 0 share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper

				_, err := keeper.ExitPool(suite.Ctx, suite.TestAccs[0], poolId, sdk.NewInt(-1), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting negative share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper

				// Test the "tokenOutMins"
				// In this case, to refund the 50000000 amount of share token, the foo, bar token are expected to be refunded as 5000 amounts.
				_, err := keeper.ExitPool(suite.Ctx, suite.TestAccs[0], poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(5001)),
				})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.App.GAMMKeeper

				// Test the "tokenOutMins"
				// In this case, to refund the 50000000 amount of share token, the foo, bar token are expected to be refunded as 5000 amounts.
				_, err := keeper.ExitPool(suite.Ctx, suite.TestAccs[0], poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(5000)),
				})
				suite.Require().NoError(err)
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range suite.TestAccs {
			suite.FundAcc(acc, defaultAcctFunds)

			// Create the pool at first
			msg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
			suite.Require().NoError(err)

			test.fn(poolId)
		}
	}
}

// TestJoinPoolExitPool_InverseRelationship tests that joining pool and exiting pool
// guarantees same amount in and out
func (suite *KeeperTestSuite) TestJoinPoolExitPool_InverseRelationship() {
	testCases := []struct {
		name             string
		pool             balancertypes.MsgCreateBalancerPool
		joinPoolShareAmt sdk.Int
	}{
		{
			name: "pool with same token ratio",
			pool: balancer.NewMsgCreateBalancerPool(nil, balancer.PoolParams{
				SwapFee: sdk.ZeroDec(),
				ExitFee: sdk.ZeroDec(),
			}, []balancertypes.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
				},
			}, defaultFutureGovernor),
			joinPoolShareAmt: types.OneShare.MulRaw(50),
		},
		{
			name: "pool with different token ratio",
			pool: balancer.NewMsgCreateBalancerPool(nil, balancer.PoolParams{
				SwapFee: sdk.ZeroDec(),
				ExitFee: sdk.ZeroDec(),
			}, []balancertypes.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(7000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
				},
			}, defaultFutureGovernor),
			joinPoolShareAmt: types.OneShare.MulRaw(50),
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()

		for _, acc := range suite.TestAccs {
			suite.FundAcc(acc, defaultAcctFunds)
		}

		createPoolAcc := suite.TestAccs[0]
		joinPoolAcc := suite.TestAccs[1]

		// test account is set on every test case iteration, we need to manually update address for pool creator
		tc.pool.Sender = createPoolAcc.String()

		poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, tc.pool)
		suite.Require().NoError(err)

		balanceBeforeJoin := suite.App.BankKeeper.GetAllBalances(suite.Ctx, joinPoolAcc)

		err = suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, joinPoolAcc, poolId, tc.joinPoolShareAmt, sdk.Coins{})
		suite.Require().NoError(err)

		_, err = suite.App.GAMMKeeper.ExitPool(suite.Ctx, joinPoolAcc, poolId, tc.joinPoolShareAmt, sdk.Coins{})

		balanceAfterExit := suite.App.BankKeeper.GetAllBalances(suite.Ctx, joinPoolAcc)
		deltaBalance, _ := balanceBeforeJoin.SafeSub(balanceAfterExit)

		// due to rounding, `balanceBeforeJoin` and `balanceAfterExit` have neglectable difference
		// coming from rounding in exitPool.Here we test if the difference is within rounding tolerance range
		roundingToleranceCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1)), sdk.NewCoin("bar", sdk.NewInt(1)))
		suite.Require().True(deltaBalance.AmountOf("foo").LTE(roundingToleranceCoins.AmountOf("foo")))
		suite.Require().True(deltaBalance.AmountOf("bar").LTE(roundingToleranceCoins.AmountOf("bar")))
	}
}

func (suite *KeeperTestSuite) TestActiveBalancerPool() {
	type testCase struct {
		blockTime  time.Time
		expectPass bool
	}

	testCases := []testCase{
		{time.Unix(1000, 0), true},
		{time.Unix(2000, 0), true},
	}

	for _, tc := range testCases {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range suite.TestAccs {
			suite.FundAcc(acc, defaultAcctFunds)

			// Create the pool at first
			poolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})
			suite.Ctx = suite.Ctx.WithBlockTime(tc.blockTime)

			// uneffected by start time
			err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, suite.TestAccs[0], poolId, types.OneShare.MulRaw(50), sdk.Coins{})
			suite.Require().NoError(err)
			_, err = suite.App.GAMMKeeper.ExitPool(suite.Ctx, suite.TestAccs[0], poolId, types.InitPoolSharesSupply.QuoRaw(2), sdk.Coins{})
			suite.Require().NoError(err)

			foocoin := sdk.NewCoin("foo", sdk.NewInt(10))
			foocoins := sdk.Coins{foocoin}

			if tc.expectPass {
				_, err = suite.App.GAMMKeeper.JoinSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], poolId, foocoins, sdk.ZeroInt())
				suite.Require().NoError(err)
				_, err = suite.App.GAMMKeeper.JoinSwapShareAmountOut(suite.Ctx, suite.TestAccs[0], poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)
				_, err = suite.App.GAMMKeeper.ExitSwapShareAmountIn(suite.Ctx, suite.TestAccs[0], poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().NoError(err)
				_, err = suite.App.GAMMKeeper.ExitSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				_, err = suite.App.GAMMKeeper.JoinSwapShareAmountOut(suite.Ctx, suite.TestAccs[0], poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().Error(err)
				_, err = suite.App.GAMMKeeper.ExitSwapShareAmountIn(suite.Ctx, suite.TestAccs[0], poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().Error(err)
				_, err = suite.App.GAMMKeeper.ExitSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().Error(err)
			}
		}
	}
}

func (suite *KeeperTestSuite) TestJoinSwapExactAmountInConsistency() {
	testCases := []struct {
		name              string
		poolSwapFee       sdk.Dec
		poolExitFee       sdk.Dec
		tokensIn          sdk.Coins
		shareOutMinAmount sdk.Int
		expectedSharesOut sdk.Int
		tokenOutMinAmount sdk.Int
	}{
		{
			name:              "single coin with zero swap and exit fees",
			poolSwapFee:       sdk.ZeroDec(),
			poolExitFee:       sdk.ZeroDec(),
			tokensIn:          sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000000))),
			shareOutMinAmount: sdk.ZeroInt(),
			expectedSharesOut: sdk.NewInt(6265857020099440400),
			tokenOutMinAmount: sdk.ZeroInt(),
		},
		// TODO: Uncomment or remove this following test case once the referenced
		// issue is resolved.
		//
		// Ref: https://github.com/osmosis-labs/osmosis/issues/1196
		// {
		// 	name:              "single coin with positive swap fee and zero exit fee",
		// 	poolSwapFee:       sdk.NewDecWithPrec(1, 2),
		// 	poolExitFee:       sdk.ZeroDec(),
		// 	tokensIn:          sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000000))),
		// 	shareOutMinAmount: sdk.ZeroInt(),
		// 	expectedSharesOut: sdk.NewInt(6226484702880621000),
		// 	tokenOutMinAmount: sdk.ZeroInt(),
		// },
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := suite.Ctx

			poolID := suite.prepareCustomBalancerPool(
				defaultAcctFunds,
				[]balancertypes.PoolAsset{
					{
						Weight: sdk.NewInt(100),
						Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
					},
					{
						Weight: sdk.NewInt(200),
						Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
					},
				},
				balancer.PoolParams{
					SwapFee: tc.poolSwapFee,
					ExitFee: tc.poolExitFee,
				},
			)

			shares, err := suite.App.GAMMKeeper.JoinSwapExactAmountIn(ctx, suite.TestAccs[0], poolID, tc.tokensIn, tc.shareOutMinAmount)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedSharesOut, shares)

			tokenOutAmt, err := suite.App.GAMMKeeper.ExitSwapShareAmountIn(
				ctx,
				suite.TestAccs[0],
				poolID,
				tc.tokensIn[0].Denom,
				shares,
				tc.tokenOutMinAmount,
			)
			suite.Require().NoError(err)

			// require swapTokenOutAmt <= (tokenInAmt * (1 - tc.poolSwapFee))
			oneMinusSwapFee := sdk.OneDec().Sub(tc.poolSwapFee)
			swapFeeAdjustedAmount := oneMinusSwapFee.MulInt(tc.tokensIn[0].Amount).RoundInt()
			suite.Require().True(tokenOutAmt.LTE(swapFeeAdjustedAmount))

			// require swapTokenOutAmt + 10 > input
			suite.Require().True(
				swapFeeAdjustedAmount.Sub(tokenOutAmt).LTE(sdk.NewInt(10)),
				"expected out amount %s, actual out amount %s",
				swapFeeAdjustedAmount, tokenOutAmt,
			)
		})
	}
}

// func (suite *KeeperTestSuite) TestSetStableSwapScalingFactors() {
// 	stableSwapPoolParams := stableswap.PoolParams{
// 		SwapFee: defaultSwapFee,
// 		ExitFee: defaultExitFee,
// 	}

// 	testPoolAsset := sdk.Coins{
// 		sdk.NewCoin("foo", sdk.NewInt(10000)),
// 		sdk.NewCoin("bar", sdk.NewInt(10000)),
// 	}

// 	suite.FundAcc(suite.TestAccs[0], defaultAcctFunds)

// 	testScalingFactors := []uint64{1, 1}

// 	msg := stableswap.NewMsgCreateStableswapPool(
// 		suite.TestAccs[0], stableSwapPoolParams, testPoolAsset, defaultFutureGovernor)
// 	poolID, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
// 	suite.Require().NoError(err)

// 	err = suite.App.GAMMKeeper.SetStableSwapScalingFactors(suite.Ctx, testScalingFactors, poolID, "")
// 	suite.Require().NoError(err)

// 	poolI, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolID)
// 	suite.Require().NoError(err)

// 	poolScalingFactors := poolI.(*stableswap.Pool).GetScalingFactors()

// 	suite.Require().Equal(
// 		poolScalingFactors,
// 		testScalingFactors,
// 	)
// }
