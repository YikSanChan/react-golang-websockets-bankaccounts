import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

const BankAccount = () => {
  const { account_id } = useParams();
  const [balance, setBalance] = useState<number | null>(null);
  const [deposit, setDeposit] = useState<number>(0);

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    const e = new EventSource(`http://localhost:8080/events/${account_id}`);
    e.onmessage = function (event) {
      console.log(
        `onmessage: got ${JSON.stringify(event, null, 2)} from ${e.url}`
      );
      fetchData();
    };
    e.onerror = function (event) {
      console.log(
        `onerror: got ${JSON.stringify(event, null, 2)} from ${e.url}`
      );
    };
  }, []);

  async function fetchData() {
    const response = await fetch(
      `http://localhost:8080/account/${account_id}/balance`
    );
    const data = await response.json(); // {"balance": 42}
    setBalance(data.balance);
  }

  // @ts-ignore
  async function handleDeposit(event) {
    const response = await fetch(
      `http://localhost:8080/account/${account_id}/deposit/${deposit}`,
      {
        method: "POST",
      }
    );
    alert("done");
    event.preventDefault();
  }

  console.log("re-render");

  return (
    <div>
      <h2>Account {account_id}</h2>
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
