# TradeInGo
Trading bot in golang that makes few trades with low risk management

# How it works
In every price chart you can find what they call resistances and supports, especially in crypto charts. Those are special levels in which the price gets trapped, once the price gets over or under one key level we can expect that it will grow or fall to the next level.

![BTCUSDT_2022-08-22_11-26-24](https://user-images.githubusercontent.com/34452508/185888008-67e893ea-5da1-46a4-a4ee-97d4681c5253.png)

## Take profit and stop loss
In order to lower the risk, before opening a position, the bot set a *takeprofit* and a *stoploss* 

- The *TAKEPROFIT* is set to the next key level the price could go.

- The *STOPLOSS* is set to the current level the price gets over/under

In this way we can limit the losses and maximize the profit!

## So how to determine those areas and key levels?
We can observe the shadows of a group of candles and see if they shares a certain price area.

![BTCUSDT_2022-08-22_11-33-36](https://user-images.githubusercontent.com/34452508/185889574-49e675f6-e8d9-4f2b-8122-1c6d83437fb7.png)

Another good indicator are the clusters. A cluster is formed when there are 2 candles going in the same direction and one in between them going in the opposite direction with a smaller body, as you can see on this screenshot


![BTCUSD_2022-10-09_16-15-07](https://user-images.githubusercontent.com/34452508/194761751-2cc0bb3c-2aeb-453c-85a4-ace27d077dc8.png)


Another way to find interesting key levels is the Fibonacci retracement. 
(See https://www.investopedia.com/ask/answers/05/fibonacciretracement.asp#:~:text=Fibonacci%20retracement%20levels%20are%20horizontal,trend%20is%20likely%20to%20continue.)

## How to start the bot
The bot can be launched using the *go run . [mode]* command. 

The mode can be: 

  - *live* for predicting real time btc price and simulating LONG or SHORT positions
  
  - *test* to run a backtest and see the performance.
  
In the repo you can find the *log.txt* file that contains the *test* output of ~ 6 months of run.
You can notice (searching for POSITION CLOSED) that the bot made few trades and only 1 of them resulted in a loss.
