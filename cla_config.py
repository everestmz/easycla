"""
Rename this file to cla_config.py and put it at the root of your CLA system application in order to
use it as your default configuration. It will override any configuration options found in
the default CLA system configuration file located at cla/config.py.
"""

DEBUG = True # Debug on for development purposes.

BASE_URL = 'https://your.cla.domain.com' # Base URL used for callbacks and OAuth2 redirects.
# Default callback once signature is completed - a custom endpoint can be provided if needed.
SIGNED_CALLBACK_URL = BASE_URL + '/v1/signed'

# Define the database we are working with.
DATABASE = 'DynamoDB' # TODO: Should use in-memory database as default.
DATABASE_HOST = 'http://localhost:50458'

# Define the key-value store used.
KEYVALUE = 'Memory'
KEYVALUE_HOST = ''

# DynamoDB-specific configurations - this is applied to each table.
DYNAMO_REGION = 'us-west-2'
DYNAMO_WRITE_UNITS = 1
DYNAMO_READ_UNITS = 1

# Define the signing service to use.
SIGNING_SERVICE = 'DocuSign'
DOCUSIGN_ROOT_URL = 'https://demo.docusign.net/restapi/v2' # This is the demo endpoint.
DOCUSIGN_USERNAME = '' # DocuSign Email or UUID.
DOCUSIGN_PASSWORD = '' # Account Password.
DOCUSIGN_INTEGRATOR_KEY = '' # Integrator Key UUID.

# GitHub Application Service.
GITHUB_APP_WEBHOOK_SECRET = '8f19e80bfe465c03ddf9dfa1b517304e00e51288'
GITHUB_APP_PRIVATE_KEY_PATH = './contributor-license-agreement.2017-08-24.private-key.pem'
GITHUB_APP_CLIENT_ID = 'Iv1.3899cde846512a7e'
GITHUB_APP_SECRET = 'c3fd6268f75752dc0dac8dc87aa64d8a8e28b633'
GITHUB_APP_ID = '4449'

# KeyCloak Authentication
KEYCLOAK_ENDPOINT = ''

# CINCO (SFDC) Authentication
CINCO_ENDPOINT = ''

# Email Service.
EMAIL_SERVICE = 'SMTP' # SMTP is useful for testing with MailHog or the like.

# SMTP Configuration.
SMTP_SENDER_EMAIL_ADDRESS = 'donotreply@cla.system'
SMTP_HOST = 'localhost'
SMTP_PORT = '25'

# Storage service.
STORAGE_SERVICE = 'LocalStorage'

# Local Storage Configuration.
LOCAL_STORAGE_FOLDER = '/tmp/cla' # Storage of uploaded agreements - change to persistent location.
