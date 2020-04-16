import React, { useEffect, useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";

const SOCKET_URL = "ws://localhost:8080/ws";
const CONNECTION_STATUSES = {
  [ReadyState.CONNECTING]: "Connecting",
  [ReadyState.OPEN]: "Open",
  [ReadyState.CLOSING]: "Closing",
  [ReadyState.CLOSED]: "Closed",
};

const BankAccount = () => {
  const [balance, setBalance] = useState<number | null>(null);
  const [deposit, setDeposit] = useState<number>(0);

  const [sendMessage, lastMessage, readyState, getWebSocket] = useWebSocket(
    SOCKET_URL
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
      const response = await fetch("http://localhost:8080/balance");
      const data = await response.json(); // {"balance": 42}
      setBalance(data.balance);
    };
    fetchData();
  }, [lastMessage]);

  // @ts-ignore
  const handleDeposit = async (event) => {
    const response = await fetch("http://localhost:8080/deposit/" + deposit, {
      method: "POST",
    });
    await response.json();
    event.preventDefault();
  };

  return (
    <div>
      <div>The WebSocket is currently {connectionStatus}</div>
      {lastMessage ? <div>Last message: {lastMessage.data}</div> : null}
      <div>Your balance is {balance || "unknown"}</div>
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

const App = () => <BankAccount />;

export default App;
