# shibtest

<h4 align="center">Brute Force Login using Shibboleth</h4>
     
<p align="center">
  <a href="#Usage">Usage</a> •
  <a href="#Installation">Installation</a> •
</p>

---
Shibtest is a simple Shibboleth login brute force tool. It essentially needs the URL which initiates the login process
from  service provider a service provider. If provided usernames and passwords to try are taken from wordlists,
otherwise the most popular ones from Daniel Miessler's SecList are used.

# Usage

```shell
shibtest -h
```
This will display help for the tool. Here are all the switches it supports.

```yaml
Shibtest is a simple Shibboleth login brute force tool. It essentially needs the URL which initiates the login process 
  from  service provider a service provider. If provided usernames and passwords to try are taken from wordlists, 
  otherwise the most popular ones from Daniel Miessler's SecList are used.   

Usage:
  shibtest [flags]

Flags:
TARGET:
  -su           the service provider start URL which initiates the authentication flow
  -iu           the identity provider base URL (everything before /idp, https://example.com)
INPUT:
  -ul           path to file containing a list of user names to use
  -u            user to use
  -pl           path to file containing a list of passwords to use
FLOWS:
  -rf           Shibboleth SAML 2.0 Redirect SSO Flow
  -pf           Shibboleth SAML 2.0 POST SSO Flow (default)
```  

# Installation

shibtest requires **go1.17** to install successfully. Run the following command to get the repo -

```sh
go install -v github.com/secinto/hacks/shibtest@latest
```
