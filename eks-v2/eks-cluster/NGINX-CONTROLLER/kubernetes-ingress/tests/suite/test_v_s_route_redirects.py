import pytest
import requests
from kubernetes.client.rest import ApiException

from settings import TEST_DATA
from suite.custom_assertions import assert_event_and_get_count, wait_and_assert_status_code, \
    assert_event_count_increased, assert_event_starts_with_text_and_contains_errors
from suite.vs_vsr_resources_utils import get_vs_nginx_template_conf, patch_v_s_route_from_yaml
from suite.resources_utils import get_first_pod_name, get_events, wait_before_test


@pytest.mark.vsr
@pytest.mark.parametrize('crd_ingress_controller, v_s_route_setup',
                         [({"type": "complete", "extra_args": [f"-enable-custom-resources"]},
                           {"example": "virtual-server-route-redirects"})],
                         indirect=True)
class TestVSRRedirects:
    def test_config(self, kube_apis, ingress_controller_prerequisites, crd_ingress_controller, v_s_route_setup):
        wait_before_test(1)
        ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
        config = get_vs_nginx_template_conf(kube_apis.v1,
                                            v_s_route_setup.namespace,
                                            v_s_route_setup.vs_name,
                                            ic_pod_name,
                                            ingress_controller_prerequisites.namespace)
        assert 'error_page 418 =307' in config and 'error_page 418 =301' in config

    def test_custom_redirect(self, kube_apis, crd_ingress_controller, v_s_route_setup):
        req_host = f"{v_s_route_setup.public_endpoint.public_ip}:{v_s_route_setup.public_endpoint.port}"
        req_url = f"http://{req_host}{v_s_route_setup.route_m.paths[0]}?arg1=arg"
        wait_and_assert_status_code(307, req_url, v_s_route_setup.vs_host, allow_redirects=False)
        resp = requests.get(req_url, headers={"host": v_s_route_setup.vs_host}, allow_redirects=False)
        assert resp.headers['location'] == "http://example.com"

    def test_default_redirect(self, kube_apis, crd_ingress_controller, v_s_route_setup):
        req_host = f"{v_s_route_setup.public_endpoint.public_ip}:{v_s_route_setup.public_endpoint.port}"
        req_url = f"http://{req_host}{v_s_route_setup.route_m.paths[1]}"
        wait_and_assert_status_code(301, req_url, v_s_route_setup.vs_host, allow_redirects=False)
        resp = requests.get(req_url, headers={"host": v_s_route_setup.vs_host}, allow_redirects=False)
        assert resp.headers['location'] == f"http://{v_s_route_setup.vs_host}/backends/default-redirect?arg="

    def test_update(self, kube_apis, crd_ingress_controller, v_s_route_setup):
        req_host = f"{v_s_route_setup.public_endpoint.public_ip}:{v_s_route_setup.public_endpoint.port}"
        req_url_1 = f"http://{req_host}{v_s_route_setup.route_m.paths[0]}"
        req_url_2 = f"http://{req_host}{v_s_route_setup.route_m.paths[1]}"
        vs_name = f"{v_s_route_setup.namespace}/{v_s_route_setup.vs_name}"
        vsr_name = f"{v_s_route_setup.namespace}/{v_s_route_setup.route_m.name}"
        vs_event_text = f"Configuration for {vs_name} was added or updated"
        vsr_event_text = f"Configuration for {vsr_name} was added or updated"
        wait_before_test(1)
        events_ns = get_events(kube_apis.v1, v_s_route_setup.namespace)
        initial_count_vs = assert_event_and_get_count(vs_event_text, events_ns)
        initial_count_vsr = assert_event_and_get_count(vsr_event_text, events_ns)
        vsr_src = f"{TEST_DATA}/virtual-server-route-redirects/route-multiple-updated.yaml"
        patch_v_s_route_from_yaml(kube_apis.custom_objects,
                                  v_s_route_setup.route_m.name, vsr_src, v_s_route_setup.namespace)
        wait_and_assert_status_code(301, req_url_1, v_s_route_setup.vs_host, allow_redirects=False)
        resp = requests.get(req_url_1, headers={"host": v_s_route_setup.vs_host}, allow_redirects=False)
        assert resp.headers['location'] == "http://demo.nginx.com"

        wait_and_assert_status_code(302, req_url_2, v_s_route_setup.vs_host, allow_redirects=False)
        resp = requests.get(req_url_2, headers={"host": v_s_route_setup.vs_host}, allow_redirects=False)
        assert resp.headers['location'] == "http://demo.nginx.com"

        new_events_ns = get_events(kube_apis.v1, v_s_route_setup.namespace)
        assert_event_count_increased(vs_event_text, initial_count_vs, new_events_ns)
        assert_event_count_increased(vsr_event_text, initial_count_vsr, new_events_ns)

    def test_validation_flow(self, kube_apis, crd_ingress_controller, v_s_route_setup):
        req_host = f"{v_s_route_setup.public_endpoint.public_ip}:{v_s_route_setup.public_endpoint.port}"
        req_url = f"http://{req_host}{v_s_route_setup.route_s.paths[0]}"
        text = f"{v_s_route_setup.namespace}/{v_s_route_setup.route_m.name}"
        event_text = f"VirtualServerRoute {text} was rejected with error:"
        invalid_fields = [
            "spec.subroutes[0].action.redirect.code", "spec.subroutes[1].action.redirect.url"
        ]
        vsr_src = f"{TEST_DATA}/virtual-server-route-redirects/route-multiple-invalid.yaml"
        patch_v_s_route_from_yaml(kube_apis.custom_objects,
                                  v_s_route_setup.route_m.name, vsr_src, v_s_route_setup.namespace)
        wait_before_test(2)

        wait_and_assert_status_code(404, req_url, v_s_route_setup.vs_host, allow_redirects=False)
        events = get_events(kube_apis.v1, v_s_route_setup.route_m.namespace)
        assert_event_starts_with_text_and_contains_errors(event_text, events, invalid_fields)

    def test_openapi_validation_flow(self, kube_apis, ingress_controller_prerequisites,
                                     crd_ingress_controller, v_s_route_setup):
        ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
        config_old = get_vs_nginx_template_conf(kube_apis.v1,
                                                v_s_route_setup.namespace,
                                                v_s_route_setup.vs_name,
                                                ic_pod_name,
                                                ingress_controller_prerequisites.namespace)
        vsr_src = f"{TEST_DATA}/virtual-server-route-redirects/route-multiple-invalid-openapi.yaml"
        try:
            patch_v_s_route_from_yaml(kube_apis.custom_objects,
                                      v_s_route_setup.route_m.name, vsr_src, v_s_route_setup.namespace)
        except ApiException as ex:
            assert ex.status == 422 \
                   and "spec.subroutes.action.redirect.url" in ex.body \
                   and "spec.subroutes.action.redirect.code" in ex.body
        except Exception as ex:
            pytest.fail(f"An unexpected exception is raised: {ex}")
        else:
            pytest.fail("Expected an exception but there was none")

        wait_before_test(1)
        config_new = get_vs_nginx_template_conf(kube_apis.v1,
                                                v_s_route_setup.namespace,
                                                v_s_route_setup.vs_name,
                                                ic_pod_name,
                                                ingress_controller_prerequisites.namespace)
        assert config_old == config_new, "Expected: config doesn't change"
