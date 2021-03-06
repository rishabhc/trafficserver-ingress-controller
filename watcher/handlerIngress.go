/*

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package watcher

import (
	//	"encoding/json"
	"log"

	"ingress-ats/endpoint"
	//	t "ingress-ats/types"
	"ingress-ats/util"

	v1beta1 "k8s.io/api/extensions/v1beta1"
)

// IgHandler implements EventHandler
type IgHandler struct {
	ResourceName string
	Ep           *endpoint.Endpoint
}

func (g *IgHandler) Add(obj interface{}) {
	log.Printf("\n\nIn INGRESS_HANDLER ADD %#v \n\n", obj)
	g.add(obj)
	g.Ep.RedisClient.PrintAllKeys()
}

func (g *IgHandler) add(obj interface{}) {
	ingressObj, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Println("In HandlerIngress Add; cannot cast to *v1beta1.Ingress")
		return
	}

	namespace := ingressObj.GetNamespace()
	ingressClass, _ := util.ExtractIngressClass(ingressObj.GetAnnotations())
	if !g.Ep.NsManager.IncludeNamespace(namespace) || !g.Ep.ATSManager.IncludeIngressClass(ingressClass) {
		log.Println("Namespace not included or Ingress Class not matched")
		return
	}

	// add the script before adding route
	snippet, snippetErr := util.ExtractServerSnippet(ingressObj.GetAnnotations())
	if snippetErr == nil {
		name := ingressObj.GetName()
		version := ingressObj.GetResourceVersion()
		nameversion := util.ConstructNameVersionString(namespace, name, version)
		g.Ep.RedisClient.DBOneSAdd(nameversion, snippet)
	}

	tlsHosts := make(map[string]string)

	for _, ingressTLS := range ingressObj.Spec.TLS {
		for _, tlsHost := range ingressTLS.Hosts {
			tlsHosts[tlsHost] = "1"
		}
	}

	for _, ingressRule := range ingressObj.Spec.Rules {
		host := ingressRule.Host
		scheme := "http"
		if _, ok := tlsHosts[host]; ok {
			scheme = "https"
		}

		for _, httpPath := range ingressRule.HTTP.Paths {

			path := httpPath.Path
			hostPath := util.ConstructHostPathString(scheme, host, path)
			service := httpPath.Backend.ServiceName
			port := httpPath.Backend.ServicePort.String()
			svcport := util.ConstructSvcPortString(namespace, service, port)

			g.Ep.RedisClient.DBOneSAdd(hostPath, svcport)

			if snippetErr == nil {
				name := ingressObj.GetName()
				version := ingressObj.GetResourceVersion()
				nameversion := util.ConstructNameVersionString(namespace, name, version)
				g.Ep.RedisClient.DBOneSAdd(hostPath, nameversion)
			}
		}

	}
}

// Update for EventHandler
func (g *IgHandler) Update(obj, newObj interface{}) {
	log.Printf("\n\nIn INGRESS_HANDLER UPDATE %#v \n\n", newObj)
	g.update(obj, newObj)
	g.Ep.RedisClient.PrintAllKeys()
}

func (g *IgHandler) update(obj, newObj interface{}) {
	ingressObj, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Println("In HandlerIngress Update; cannot cast to *v1beta1.Ingress")
		return
	}

	newIngressObj, ok := newObj.(*v1beta1.Ingress)
	if !ok {
		log.Println("In HandlerIngress Update; cannot cast to *v1beta1.Ingress")
		return
	}

	m := make(map[string]string)

	namespace := ingressObj.GetNamespace()
	ingressClass, _ := util.ExtractIngressClass(ingressObj.GetAnnotations())
	if g.Ep.NsManager.IncludeNamespace(namespace) && g.Ep.ATSManager.IncludeIngressClass(ingressClass) {
		log.Println("Old Namespace included")

		_, snippetErr := util.ExtractServerSnippet(ingressObj.GetAnnotations())

		tlsHosts := make(map[string]string)

		for _, ingressTLS := range ingressObj.Spec.TLS {
			for _, tlsHost := range ingressTLS.Hosts {
				tlsHosts[tlsHost] = "1"
			}
		}

		for _, ingressRule := range ingressObj.Spec.Rules {
			host := ingressRule.Host
			scheme := "http"
			if _, ok := tlsHosts[host]; ok {
				scheme = "https"
			}

			for _, httpPath := range ingressRule.HTTP.Paths {

				path := httpPath.Path
				hostPath := util.ConstructHostPathString(scheme, host, path)

				g.Ep.RedisClient.DBOneSUnionStore("temp_"+hostPath, hostPath)
				m["temp_"+hostPath] = hostPath

				service := httpPath.Backend.ServiceName
				port := httpPath.Backend.ServicePort.String()
				svcport := util.ConstructSvcPortString(namespace, service, port)

				g.Ep.RedisClient.DBOneSRem("temp_"+hostPath, svcport)

				if snippetErr == nil {
					name := ingressObj.GetName()
					version := ingressObj.GetResourceVersion()
					nameversion := util.ConstructNameVersionString(namespace, name, version)
					g.Ep.RedisClient.DBOneSRem("temp_"+hostPath, nameversion)
				}
			}

		}
	}

	newNamespace := ingressObj.GetNamespace()
	newIngressClass, _ := util.ExtractIngressClass(ingressObj.GetAnnotations())
	if g.Ep.NsManager.IncludeNamespace(newNamespace) && g.Ep.ATSManager.IncludeIngressClass(newIngressClass) {
		log.Println("New Namespace included")

		newSnippet, newSnippetErr := util.ExtractServerSnippet(newIngressObj.GetAnnotations())
		if newSnippetErr == nil {
			newName := newIngressObj.GetName()
			newVersion := newIngressObj.GetResourceVersion()
			newNameversion := util.ConstructNameVersionString(newNamespace, newName, newVersion)
			g.Ep.RedisClient.DBOneSAdd(newNameversion, newSnippet)
		}

		newTlsHosts := make(map[string]string)

		for _, newIngressTLS := range newIngressObj.Spec.TLS {
			for _, newTlsHost := range newIngressTLS.Hosts {
				newTlsHosts[newTlsHost] = "1"
			}
		}

		for _, ingressRule := range newIngressObj.Spec.Rules {
			host := ingressRule.Host
			scheme := "http"
			if _, ok := newTlsHosts[host]; ok {
				scheme = "https"
			}

			for _, httpPath := range ingressRule.HTTP.Paths {

				path := httpPath.Path
				hostPath := util.ConstructHostPathString(scheme, host, path)

				service := httpPath.Backend.ServiceName
				port := httpPath.Backend.ServicePort.String()
				svcport := util.ConstructSvcPortString(namespace, service, port)

				g.Ep.RedisClient.DBOneSAdd("temp_"+hostPath, svcport)
				m["temp_"+hostPath] = hostPath

				if newSnippetErr == nil {
					newName := newIngressObj.GetName()
					newVersion := newIngressObj.GetResourceVersion()
					newNameversion := util.ConstructNameVersionString(newNamespace, newName, newVersion)
					g.Ep.RedisClient.DBOneSAdd("temp_"+hostPath, newNameversion)
				}
			}

		}
	}

	for key, value := range m {
		g.Ep.RedisClient.DBOneSUnionStore(value, key)
		g.Ep.RedisClient.DBOneDel(key)
	}
}

// Delete for EventHandler
func (g *IgHandler) Delete(obj interface{}) {
	log.Printf("\n\nIn INGRESS_HANDLER DELETE %#v \n\n", obj)
	g.delete(obj)
	g.Ep.RedisClient.PrintAllKeys()
}

// Helper for Deletes
func (g *IgHandler) delete(obj interface{}) {
	ingressObj, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Println("In HandlerIngress Delete; cannot cast to *v1beta1.Ingress")
		return
	}

	namespace := ingressObj.GetNamespace()
	ingressClass, _ := util.ExtractIngressClass(ingressObj.GetAnnotations())
	if !g.Ep.NsManager.IncludeNamespace(namespace) || !g.Ep.ATSManager.IncludeIngressClass(ingressClass) {
		log.Println("Namespace not included or Ingress Class not matched")
		return
	}

	_, snippetErr := util.ExtractServerSnippet(ingressObj.GetAnnotations())

	tlsHosts := make(map[string]string)

	for _, ingressTLS := range ingressObj.Spec.TLS {
		for _, tlsHost := range ingressTLS.Hosts {
			tlsHosts[tlsHost] = "1"
		}
	}

	for _, ingressRule := range ingressObj.Spec.Rules {
		host := ingressRule.Host
		scheme := "http"
		if _, ok := tlsHosts[host]; ok {
			scheme = "https"
		}

		for _, httpPath := range ingressRule.HTTP.Paths {

			path := httpPath.Path
			hostPath := util.ConstructHostPathString(scheme, host, path)
			service := httpPath.Backend.ServiceName
			port := httpPath.Backend.ServicePort.String()
			svcport := util.ConstructSvcPortString(namespace, service, port)

			g.Ep.RedisClient.DBOneSRem(hostPath, svcport)

			if snippetErr == nil {
				name := ingressObj.GetName()
				version := ingressObj.GetResourceVersion()
				nameversion := util.ConstructNameVersionString(namespace, name, version)
				g.Ep.RedisClient.DBOneSRem(hostPath, nameversion)
			}
		}

	}
}

// GetResourceName returns the resource name
func (g *IgHandler) GetResourceName() string {
	return g.ResourceName
}
