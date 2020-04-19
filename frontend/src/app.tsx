import React from "react";
import BankAccount from "./bank-account";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";

const App = () => (
  <Router>
    <div>
      <h1>Bank</h1>
      <Switch>
        <Route path="/:account_id">
          <BankAccount />
        </Route>
      </Switch>
    </div>
  </Router>
);

export default App;
