const { Provider } = require('oidc-provider');

const configuration = {
  clients: [{
    client_id: 'dev-client',
    client_secret: 'secret',
    redirect_uris: ['http://localhost:8080/api/v1/auth/oidc/callback'],
    response_types: ['code'],
    grant_types: ['authorization_code'],
  }],
};

const oidc = new Provider('http://localhost:4000', configuration);

oidc.listen(4000, () => {
  console.log('OIDC provider running on http://localhost:4000');
});
