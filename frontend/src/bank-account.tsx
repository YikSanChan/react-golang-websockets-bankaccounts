import React, { useEffect, useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";
import { useParams } from "react-router-dom";

const SOCKET_URL = "ws://localhost:8080/ws";
const CONNECTION_STATUSES = {
  [ReadyState.CONNECTING]: "Connecting",
  [ReadyState.OPEN]: "Open",
  [ReadyState.CLOSING]: "Closing",
  [ReadyState.CLOSED]: "Closed",
};

const BankAccount = () => {
  const { account_id } = useParams();
  const [balance, setBalance] = useState<number | null>(null);
  const [deposit, setDeposit] = useState<number>(0);

  const [sendMessage, lastMessage, readyState, getWebSocket] = useWebSocket(
    `${SOCKET_URL}/${account_id}`
  );

  const connectionStatus = CONNECTION_STATUSES[readyState];

  useEffect(() => {
    if (lastMessage !== null) {
      // getWebSocket returns the WebSocket wrapped in a Proxy.
      // This is to restrict actions like mutating a shared websocket, overwriting handlers, etc
      const currentWebsocketUrl = getWebSocket().url;
      console.log("received a message from ", currentWebsocketUrl);
    }
  }, [lastMessage]);

  useEffect(() => {
    const fetchData = async () => {
      const response = await fetch(
        `http://localhost:8080/account/${account_id}/balance`
      );
      const data = await response.json(); // {"balance": 42}
      setBalance(data.balance);
    };
    fetchData();
  }, [lastMessage]);

  // @ts-ignore
  const handleDeposit = async (event) => {
    const response = await fetch(
      `http://localhost:8080/account/${account_id}/deposit/${deposit}`,
      {
        method: "POST",
      }
    );
    await response.json();
    event.preventDefault();
  };

  console.log("re-render");

  return (
    <div>
      <h2>Account {account_id}</h2>
      <div>The WebSocket is currently {connectionStatus}</div>
      {lastMessage ? <div>Last message: {lastMessage.data}</div> : null}
      <div>Your balance is {balance === null ? "unknown" : balance}</div>
      <form onSubmit={handleDeposit}>
        <label>
          Deposit:
          <input
            type="text"
            value={deposit}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
              setDeposit(+event.target.value)
            }
          />
        </label>
        <input type="submit" value="Submit" />
      </form>
    </div>
  );
};

export default BankAccount;
