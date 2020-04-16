import React, { useEffect, useState } from "react";

const App = () => <BankAccount />;

const BankAccount = () => {
  const [balance, setBalance] = useState<number | null>(null);
  const [deposit, setDeposit] = useState<number>(0);

  useEffect(() => {
    const fetchData = async () => {
      const response = await fetch("http://localhost:8080/balance");
      const data = await response.json(); // {"balance": 42}
      setBalance(data.balance);
    };
    fetchData();
  }, []);

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

export default App;
