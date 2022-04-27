import requests
import pytest
from kubernetes.client.rest import ApiException

from settings import TEST_DATA
from suite.custom_assertions import assert_event_and_get_count, assert_event_count_increased, assert_response_codes, \
    assert_event, assert_no_new_events
from suite.vs_vsr_resources_utils import get_vs_nginx_template_conf, patch_virtual_server_from_yaml
from suite.resources_utils import get_first_pod_name, wait_before_test, get_events


@pytest.mark.vs
@pytest.mark.parametrize('crd_ingress_controller, virtual_server_setup',
                         [({"type": "complete", "extra_args": [f"-enable-custom-resources"]},
                           {"example": "virtual-server-upstream-tls", "app_type": "secure"})],
                         indirect=True)
class TestVirtualServerUpstreamTls:
    def test_responses_and_config_after_setup(self, kube_apis, ingress_controller_prerequisites,
                                              crd_ingress_controller, virtual_server_setup):
        ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
        config = get_vs_nginx_template_conf(kube_apis.v1,
                                            virtual_server_setup.namespace,
                                            virtual_server_setup.vs_name,
                                            ic_pod_name,
                                            ingress_controller_prerequisites.namespace)
        resp_1 = requests.get(virtual_server_setup.backend_1_url,
                              headers={"host": virtual_server_setup.vs_host})
        resp_2 = requests.get(virtual_server_setup.backend_2_url,
                              headers={"host": virtual_server_setup.vs_host})

        proxy_host = f"vs_{virtual_server_setup.namespace}_{virtual_server_setup.vs_name}"
        assert f'proxy_pass https://{proxy_host}_backend1' not in config
        assert f'proxy_pass https://{proxy_host}_backend2' in config
        assert_response_codes(resp_1, resp_2)

    def test_event_after_setup(self, kube_apis, ingress_controller_prerequisites,
                               crd_ingress_controller, virtual_server_setup):
        text = f"{virtual_server_setup.namespace}/{virtual_server_setup.vs_name}"
        vs_event_text = f"Configuration for {text} was added or updated"
        events_vs = get_events(kube_apis.v1, virtual_server_setup.namespace)
        assert_event(vs_event_text, events_vs)

    def test_validation_flow(self, kube_apis, ingress_controller_prerequisites,
                             crd_ingress_controller, virtual_server_setup):
        ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
        initial_events_vs = get_events(kube_apis.v1, virtual_server_setup.namespace)
        try:
            patch_virtual_server_from_yaml(kube_apis.custom_objects,
                                           virtual_server_setup.vs_name,
                                           f"{TEST_DATA}/virtual-server-upstream-tls/virtual-server-invalid.yaml",
                                           virtual_server_setup.namespace)
        except ApiException as ex:
            assert ex.status == 422 and "spec.upstreams.tls.enable" in ex.body
        except Exception as ex:
            pytest.fail(f"An unexpected exception is raised: {ex}")
        else:
            pytest.fail("Expected an exception but there was none")

        wait_before_test(1)
        config = get_vs_nginx_template_conf(kube_apis.v1,
                                            virtual_server_setup.namespace,
                                            virtual_server_setup.vs_name,
                                            ic_pod_name,
                                            ingress_controller_prerequisites.namespace)
        resp_1 = requests.get(virtual_server_setup.backend_1_url,
                              headers={"host": virtual_server_setup.vs_host})
        resp_2 = requests.get(virtual_server_setup.backend_2_url,
                              headers={"host": virtual_server_setup.vs_host})
        new_events_vs = get_events(kube_apis.v1, virtual_server_setup.namespace)

        proxy_host = f"vs_{virtual_server_setup.namespace}_{virtual_server_setup.vs_name}"
        assert f'proxy_pass https://{proxy_host}_backend1' not in config
        assert f'proxy_pass https://{proxy_host}_backend2' in config
        assert_response_codes(resp_1, resp_2)
        assert_no_new_events(initial_events_vs, new_events_vs)

    def test_responses_and_config_after_disable_tls(self, kube_apis, ingress_controller_prerequisites,
                                                    crd_ingress_controller, virtual_server_setup):
        ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
        text = f"{virtual_server_setup.namespace}/{virtual_server_setup.vs_name}"
        vs_event_text = f"Configuration for {text} was added or updated"
        initial_events_vs = get_events(kube_apis.v1, virtual_server_setup.namespace)
        initial_count = assert_event_and_get_count(vs_event_text, initial_events_vs)
        patch_virtual_server_from_yaml(kube_apis.custom_objects,
                                       virtual_server_setup.vs_name,
                                       f"{TEST_DATA}/virtual-server-upstream-tls/virtual-server-disable-tls.yaml",
                                       virtual_server_setup.namespace)
        wait_before_test(1)
        config = get_vs_nginx_template_conf(kube_apis.v1,
                                            virtual_server_setup.namespace,
                                            virtual_server_setup.vs_name,
                                            ic_pod_name,
                                            ingress_controller_prerequisites.namespace)
        resp_1 = requests.get(virtual_server_setup.backend_1_url,
                              headers={"host": virtual_server_setup.vs_host})
        resp_2 = requests.get(virtual_server_setup.backend_2_url,
                              headers={"host": virtual_server_setup.vs_host})
        new_events_vs = get_events(kube_apis.v1, virtual_server_setup.namespace)

        assert 'proxy_pass https://' not in config
        assert_response_codes(resp_1, resp_2, 200, 400)
        assert_event_count_increased(vs_event_text, initial_count, new_events_vs)
