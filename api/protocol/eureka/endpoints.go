package eureka

// ApplicationTemplateURL returns URL path used for maintaining application instance
func applicationTemplateURL() string {
	return appTemplate
}

// ApplicationURL returns (client side) URL path used for creating new instance registrations
func ApplicationURL(token, appid string) string {
	return apiPath + "/" + token + apiVer + "/apps/" + appid
}

// InstanceURL returns (client side) URL path used for interacting with the specified instance
func InstanceURL(token, appid, id string) string {
	return apiPath + "/" + token + apiVer + "/apps/" + appid + "/" + id
}

// InstanceTemplateURL returns the router's (server side) URL path used for interacting with an instance
func instanceTemplateURL() string {
	return instanceTemplate
}

// ApplicationsURL returns URL path used for querying service names
func applicationsURL() string {
	return appsPath + "/"
}

// InstanceQueryTemplateURL returns the router's (server side) URL path used for querying instance by id
func instanceQueryTemplateURL() string {
	return instanceQueryTemplate
}

// InstanceStatusTemplateURL returns the router's (server side) URL path used for interacting with an instance status
func instanceStatusTemplateURL() string {
	return instanceStatusTemplate
}

// InstanceStatusURL returns (client side) URL path used for setting instance status
func InstanceStatusURL(token, appid, id string) string {
	return apiPath + "/" + token + apiVer + "/apps/" + appid + "/" + id + "/status"
}

// VipTemplateURL returns URL path used for for querying instances by vip address
func vipTemplateURL() string {
	return vipTemplate
}

// Eureka API parameter names
const (
	RouteParamToken      = "token"
	RouteParamAppID      = "appid"
	RouteParamInstanceID = "iid"
	RouterParamVip       = "vip"
)

const ( // Eureka API related constants
	apiPath                = "/api/eureka"
	apiVer                 = "/v2"
	tokenTemplate          = apiPath + "/#" + RouteParamToken
	appsPath               = tokenTemplate + apiVer + "/apps"
	appTemplate            = appsPath + "/#" + RouteParamAppID
	instanceTemplate       = appTemplate + "/#" + RouteParamInstanceID
	instanceQueryTemplate  = tokenTemplate + apiVer + "/instances/#" + RouteParamInstanceID
	instanceStatusTemplate = instanceTemplate + "/status"
	vipTemplate            = tokenTemplate + apiVer + "/vips/#" + RouterParamVip
)
