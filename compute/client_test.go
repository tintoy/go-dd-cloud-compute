package compute

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Get user account details (successful).
func TestClient_GetAccount_Success(test *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/xml")
		writer.WriteHeader(http.StatusOK)

		fmt.Fprintln(writer, accountTestResponse)
	}))
	defer testServer.Close()

	client := NewClient("au1", "user1", "password")
	client.setBaseAddress(testServer.URL)

	account, err := client.GetAccount()
	if err != nil {
		test.Fatal(err)
	}

	verifyAccountTestResponse(test, account)
}

// Get user account details (access denied).
func TestClient_GetAccount_AccessDenied(test *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		http.Error(writer, "Invalid credentials.", http.StatusUnauthorized)
	}))
	defer testServer.Close()

	client := NewClient("au1", "user", "password")
	client.setBaseAddress(testServer.URL)

	_, err := client.GetAccount()
	if err == nil {
		test.Fatal("Client did not return expected access-denied error.")

		return
	}
	if err.Error() != "Cannot connect to compute API (invalid credentials)." {
		test.Fatal("Unexpected error: ", err)
	}
}

// List network domains (successful).
func TestClient_ListNetworkDomains_Success(test *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)

		fmt.Fprintln(writer, networkDomainsTestResponse)
	}))
	defer testServer.Close()

	client := NewClient("au1", "user1", "password")
	client.setBaseAddress(testServer.URL)
	client.setAccount(&Account{
		OrganizationID: "dummy-organization-id",
	})

	networkDomains, err := client.ListNetworkDomains()
	if err != nil {
		test.Fatal(err)
	}

	verifyNetworkDomainsTestResponse(test, networkDomains)
}

/*
 * Test helpers
 */

// Configure the Client to use the specified base address.
func (client *Client) setBaseAddress(baseAddress string) error {
	if len(baseAddress) == 0 {
		return fmt.Errorf("Must supply a valid base URI.")
	}

	client.baseAddress = baseAddress

	return nil
}

// Pre-cache account details for the client.
func (client *Client) setAccount(account *Account) {
	client.stateLock.Lock()
	defer client.stateLock.Unlock()

	client.account = account
}

/*
 * Test responses.
 */

var accountTestResponse = `
    <?xml version="1.0" encoding="UTF-8" standalone="yes"?>
        <ns3:Account xmlns="http://oec.api.opsource.net/schemas/organization" xmlns:ns2="http://oec.api.opsource.net/schemas/admin" xmlns:ns4="http://oec.api.opsource.net/schemas/server" xmlns:ns3="http://oec.api.opsource.net/schemas/directory" xmlns:ns6="http://oec.api.opsource.net/schemas/datacenter" xmlns:ns5="http://oec.api.opsource.net/schemas/whitelabel" xmlns:ns8="http://oec.api.opsource.net/schemas/backup" xmlns:ns7="http://oec.api.opsource.net/schemas/general" xmlns:ns13="http://oec.api.opsource.net/schemas/serverbootstrap" xmlns:ns9="http://oec.api.opsource.net/schemas/storage" xmlns:ns12="http://oec.api.opsource.net/schemas/vip" xmlns:ns11="http://oec.api.opsource.net/schemas/network" xmlns:ns10="http://oec.api.opsource.net/schemas/manualimport" xmlns:ns16="http://oec.api.opsource.net/schemas/reset" xmlns:ns15="http://oec.api.opsource.net/schemas/multigeo" xmlns:ns14="http://oec.api.opsource.net/schemas/support">
            <ns3:userName>user1</ns3:userName>
            <ns3:fullName>User One</ns3:fullName>
            <ns3:firstName>User</ns3:firstName>
            <ns3:lastName>One</ns3:lastName>
            <ns3:emailAddress>user1@corp.com</ns3:emailAddress>
            <ns3:department>Some Department</ns3:department>
            <ns3:customDefined1></ns3:customDefined1>
            <ns3:customDefined2></ns3:customDefined2>
            <ns3:orgId>cc309bfe-1234-43b7-a6a6-2b7a1965cf63</ns3:orgId>
            <ns3:roles>
                <ns3:role>
                    <ns3:name>server</ns3:name>
                </ns3:role>
                <ns3:role>
                    <ns3:name>network</ns3:name>
                </ns3:role>
                <ns3:role>
                    <ns3:name>create image</ns3:name>
                </ns3:role>
            </ns3:roles>
        </ns3:Account>
`

func verifyAccountTestResponse(test *testing.T, account *Account) {
	expect := expect(test)

	expect.notNil("Account", account)

	expect.equalsString("Account.UserName", "user1", account.UserName)
	expect.equalsString("Account.FullName", "User One", account.FullName)
	expect.equalsString("Account.FirstName", "User", account.FirstName)
	expect.equalsString("Account.LastName", "One", account.LastName)
	expect.equalsString("Account.Department", "Some Department", account.Department)
	expect.equalsString("Account.EmailAddress", "user1@corp.com", account.EmailAddress)
	expect.equalsString("Account.OrganizationID", "cc309bfe-1234-43b7-a6a6-2b7a1965cf63", account.OrganizationID)

	expect.notNil("Account.AssignedRoles", account.AssignedRoles)
	expect.equalsInt("Account.AssignedRoles size", 3, len(account.AssignedRoles))

	role1 := account.AssignedRoles[0]
	expect.equalsString("AssignedRoles[0].Name", "server", role1.Name)

	role2 := account.AssignedRoles[1]
	expect.equalsString("AssignedRoles[1].Name", "network", role2.Name)

	role3 := account.AssignedRoles[2]
	expect.equalsString("AssignedRoles[2].Name", "create image", role3.Name)
}

var networkDomainsTestResponse = `
{
	  "networkDomain": [
	    {
	      "name": "Domain 1",
	      "description": "This is test domain 1",
	      "type": "ESSENTIALS",
	      "snatIpv4Address": "168.128.17.63",
	      "createTime": "2016-01-12T22:33:05.000Z",
	      "state": "NORMAL",
	      "id": "75ab2a57-b75e-4ec6-945a-e8c60164fdf6",
	      "datacenterId": "AU9"
	    },
	    {
	      "name": "Domain 2",
	      "description": "",
	      "type": "ESSENTIALS",
	      "snatIpv4Address": "168.128.7.18",
	      "createTime": "2016-01-18T08:56:16.000Z",
	      "state": "NORMAL",
	      "id": "b91e0ba4-322c-32ca-bbc7-50b9a72d5f98",
	      "datacenterId": "AU10"
	    }
	  ],
	  "pageNumber": 1,
	  "pageCount": 2,
	  "totalCount": 2,
	  "pageSize": 250
	}
`

func verifyNetworkDomainsTestResponse(test *testing.T, networkDomains *NetworkDomains) {
	expect := expect(test)

	expect.notNil("NetworkDomains", networkDomains)

	expect.equalsInt("NetworkDomains.PageCount", 2, networkDomains.PageCount)
	expect.equalsInt("NetworkDomains.Domains size", 2, len(networkDomains.Domains))

	domain1 := networkDomains.Domains[0]
	expect.equalsString("NetworkDomains.Domains[0].Name", "Domain 1", domain1.Name)

	domain2 := networkDomains.Domains[1]
	expect.equalsString("NetworkDomains.Domains[1].Name", "Domain 2", domain2.Name)
}
