// This code is meant to be run as an Auth0 action. https://auth0.com/docs/customize/actions

const axios = require('axios'); 

/**
* Handler that will be called during the execution of a PostLogin flow.
*
* @param {Event} event - Details about the user and the context in which they are logging in.
* @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
*/
exports.onExecutePostLogin = async (event, api) => {

  console.log(`=== new login attempt ===`);
  if (event.transaction){
    console.log(`calling ${event.transaction.redirect_uri}/can-access`);
    const response = await axios.post(`${event.transaction.redirect_uri}/can-access`,
      { email: event.user.email },
      { headers: { 'Content-Type': 'application/json' } });

    console.log(`called can-access: ${JSON.stringify(response.data)}`);

    if (response.data.allowed) {
      console.log(`Access to ${event.client.name} is allowed`);
      return;
    }
  }
  
  api.access.deny(`Access to ${event.client.name} is not allowed`);
};


/**
* Handler that will be invoked when this action is resuming after an external redirect. If your
* onExecutePostLogin function does not perform a redirect, this function can be safely ignored.
*
* @param {Event} event - Details about the user and the context in which they are logging in.
* @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
*/
// exports.onContinuePostLogin = async (event, api) => {
// };
