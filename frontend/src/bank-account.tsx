import React, { useEffect, useMemo, useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";
import { useParams } from "react-router-dom";

const SOCKET_URL = "ws://localhost:8080/subscribe";
const CONNECTION_STATUSES = {
  [ReadyState.CONNECTING]: "Connecting",
  [ReadyState.OPEN]: "Open",
  [ReadyState.CLOSING]: "Closing",
  [ReadyState.CLOSED]: "Closed",
  [ReadyState.UNINSTANTIATED]: "Uninstantiated",
};
const FIXED_DEPOSIT = 10;

const BankAccount = () => {
  const { account_id } = useParams();
  const [balance, setBalance] = useState<number | null>(null);

  const STATIC_OPTIONS = useMemo(
    () => ({
      share: true,
      filter: (message: any) => {
        console.log("Received message: " + message.data);
        const data = JSON.parse(message.data);
        return data.account_id === account_id;
      },
    }),
    []
  );

  const [sendMessage, lastMessage, readyState, getWebSocket] = useWebSocket(
    SOCKET_URL,
    STATIC_OPTIONS
  );

  const connectionStatus = CONNECTION_STATUSES[readyState];

  useEffect(() => {
    if (lastMessage !== null) {
      const data = JSON.parse(lastMessage.data);
      setBalance(data.balance);
    }
  }, [lastMessage]);

  useEffect(() => {
    async function fetchData() {
      const response = await fetch(
        `http://localhost:8080/account/${account_id}/balance`
      );
      const { balance } = await response.json(); // {"balance": 42}
      setBalance(balance);
    }
    fetchData();
  }, []);

  // @ts-ignore
  async function handleDeposit() {
    const response = await fetch(
      `http://localhost:8080/account/${account_id}/deposit/${FIXED_DEPOSIT}`,
      {
        method: "POST",
      }
    );
    const { balance } = await response.json();
    setBalance(balance);
  }

  return (
    <div>
      <h2>Account {account_id}</h2>
      <div>The WebSocket is currently {connectionStatus}</div>
      {lastMessage ? <div>Last message: {lastMessage.data}</div> : null}
      <div>Your balance is {balance === null ? "unknown" : balance}</div>
      <button onClick={handleDeposit}>deposit $10</button>
    </div>
  );
};

export default BankAccount;
