import requests, logging
import pytest, json

from settings import TEST_DATA, DEPLOYMENTS
from suite.resources_utils import (
    wait_before_test,
    create_items_from_yaml,
    wait_before_test,
    get_file_contents,
    get_service_endpoint,
)
from suite.custom_resources_utils import (
    create_crd_from_yaml,
    delete_crd,
)
from suite.vs_vsr_resources_utils import(
    delete_virtual_server,
    create_virtual_server_from_yaml,
    patch_virtual_server_from_yaml,
    patch_v_s_route_from_yaml,
    create_v_s_route_from_yaml,
    delete_v_s_route,
)
from suite.policy_resources_utils import(
    create_policy_from_yaml,
    delete_policy,
    read_policy,
)
from suite.ap_resources_utils import (
    create_ap_usersig_from_yaml,
    delete_ap_usersig,
    delete_and_create_ap_policy_from_yaml,
    read_ap_custom_resource,
    create_ap_logconf_from_yaml,
    create_ap_policy_from_yaml,
    delete_ap_policy,
    delete_ap_logconf,
    create_ap_waf_policy_from_yaml,
)
from suite.yaml_utils import get_first_ingress_host_from_yaml, get_name_from_yaml

ap_pol_name = ""
log_name = ""
std_vs_src = f"{TEST_DATA}/ap-waf/standard/virtual-server.yaml"
waf_spec_vs_src = f"{TEST_DATA}/ap-waf/virtual-server-waf-spec.yaml"
waf_route_vs_src = f"{TEST_DATA}/ap-waf/virtual-server-waf-route.yaml"
waf_subroute_vsr_src = f"{TEST_DATA}/ap-waf/virtual-server-route-waf-subroute.yaml"
waf_pol_default_src = f"{TEST_DATA}/ap-waf/policies/waf-default.yaml"
waf_pol_dataguard_src = f"{TEST_DATA}/ap-waf/policies/waf-dataguard.yaml"
ap_policy_uds = "dataguard-alarm-uds"
uds_crd_resource = f"{TEST_DATA}/ap-waf/ap-ic-uds.yaml"
valid_resp_addr = "Server address:"
valid_resp_name = "Server name:"
invalid_resp_title = "Request Rejected"
invalid_resp_body = "The requested URL was rejected. Please consult with your administrator."


@pytest.fixture(scope="class")
def appprotect_setup(request, kube_apis, test_namespace) -> None:
    """
    Deploy simple application and all the AppProtect(dataguard-alarm) resources under test in one namespace.

    :param request: pytest fixture
    :param kube_apis: client apis
    :param ingress_controller_endpoint: public endpoint
    :param test_namespace:
    """

    print("------------------------- Deploy logconf -----------------------------")
    src_log_yaml = f"{TEST_DATA}/ap-waf/logconf.yaml"
    global log_name
    log_name = create_ap_logconf_from_yaml(kube_apis.custom_objects, src_log_yaml, test_namespace)

    print("------------------------- Create UserSig CRD resource-----------------------------")
    usersig_name = create_ap_usersig_from_yaml(
        kube_apis.custom_objects, uds_crd_resource, test_namespace
    )

    print(f"------------------------- Deploy dataguard-alarm appolicy ---------------------------")
    src_pol_yaml = f"{TEST_DATA}/ap-waf/{ap_policy_uds}.yaml"
    global ap_pol_name
    ap_pol_name = create_ap_policy_from_yaml(kube_apis.custom_objects, src_pol_yaml, test_namespace)

    def fin():
        print("Clean up:")
        delete_ap_policy(kube_apis.custom_objects, ap_pol_name, test_namespace)
        delete_ap_usersig(kube_apis.custom_objects, usersig_name, test_namespace)
        delete_ap_logconf(kube_apis.custom_objects, log_name, test_namespace)

    request.addfinalizer(fin)


def assert_ap_crd_info(ap_crd_info, policy_name) -> None:
    """
    Assert fields in AppProtect policy documents
    :param ap_crd_info: CRD output from k8s API
    :param policy_name:
    """
    assert ap_crd_info["kind"] == "APPolicy"
    assert ap_crd_info["metadata"]["name"] == policy_name
    assert ap_crd_info["spec"]["policy"]["enforcementMode"] == "blocking"
    assert (
        ap_crd_info["spec"]["policy"]["blocking-settings"]["violations"][0]["name"]
        == "VIOL_DATA_GUARD"
    )


def assert_invalid_responses(response) -> None:
    """
    Assert responses when policy config is blocking requests
    :param response: Response
    """
    assert invalid_resp_title in response.text
    assert invalid_resp_body in response.text
    assert response.status_code == 200


def assert_valid_responses(response) -> None:
    """
    Assert responses when policy config is allowing requests
    :param response: Response
    """
    assert valid_resp_name in response.text
    assert valid_resp_addr in response.text
    assert response.status_code == 200


@pytest.mark.skip_for_nginx_oss
@pytest.mark.appprotect
@pytest.mark.parametrize(
    "crd_ingress_controller_with_ap, virtual_server_setup",
    [
        (
            {
                "type": "complete",
                "extra_args": [
                    f"-enable-custom-resources",
                    f"-enable-leader-election=false",
                    f"-enable-app-protect",
                ],
            },
            {"example": "ap-waf", "app_type": "simple",},
        )
    ],
    indirect=True,
)
class TestAppProtectWAFPolicyVS:
    def restore_default_vs(self, kube_apis, virtual_server_setup) -> None:
        """
        Restore VirtualServer without policy spec
        """
        delete_virtual_server(
            kube_apis.custom_objects, virtual_server_setup.vs_name, virtual_server_setup.namespace
        )
        create_virtual_server_from_yaml(
            kube_apis.custom_objects, std_vs_src, virtual_server_setup.namespace
        )
        wait_before_test()

    @pytest.mark.smoke
    @pytest.mark.parametrize(
        "vs_src, waf",
        [
            (waf_spec_vs_src, waf_pol_default_src),
            (waf_spec_vs_src, waf_pol_dataguard_src),
            (waf_route_vs_src, waf_pol_default_src),
            (waf_route_vs_src, waf_pol_dataguard_src),
        ],
    )
    def test_ap_waf_policy_block(
        self,
        kube_apis,
        crd_ingress_controller_with_ap,
        virtual_server_setup,
        appprotect_setup,
        test_namespace,
        vs_src,
        waf,
    ):
        """
        Test waf policy when enabled with default and dataguard-alarm AP Policies
        """
        print(f"Create waf policy")
        if waf == waf_pol_dataguard_src:
            create_ap_waf_policy_from_yaml(
                kube_apis.custom_objects,
                waf,
                test_namespace,
                test_namespace,
                True,
                False,
                ap_pol_name,
                log_name,
                "syslog:server=127.0.0.1:514",
            )
        elif waf == waf_pol_default_src:
            pol_name = create_policy_from_yaml(kube_apis.custom_objects, waf, test_namespace)
        else:
            pytest.fail(f"Invalid argument")

        wait_before_test()
        print(f"Patch vs with policy: {vs_src}")
        patch_virtual_server_from_yaml(
            kube_apis.custom_objects,
            virtual_server_setup.vs_name,
            vs_src,
            virtual_server_setup.namespace,
        )
        wait_before_test()
        ap_crd_info = read_ap_custom_resource(
            kube_apis.custom_objects, test_namespace, "appolicies", ap_policy_uds
        )
        assert_ap_crd_info(ap_crd_info, ap_policy_uds)
        wait_before_test(120)

        print(
            "----------------------- Send request with embedded malicious script----------------------"
        )
        response1 = requests.get(
            virtual_server_setup.backend_1_url + "</script>",
            headers={"host": virtual_server_setup.vs_host},
        )
        print(response1.text)

        print(
            "----------------------- Send request with blocked keyword in UDS----------------------"
        )
        response2 = requests.get(
            virtual_server_setup.backend_1_url,
            headers={"host": virtual_server_setup.vs_host},
            data="kic",
        )
        print(response2.text)

        delete_policy(kube_apis.custom_objects, "waf-policy", test_namespace)
        self.restore_default_vs(kube_apis, virtual_server_setup)
        assert_invalid_responses(response1)
        if waf == waf_pol_dataguard_src:
            assert_invalid_responses(response2)
        elif waf == waf_pol_default_src:
            assert_valid_responses(response2)
        else:
            pytest.fail(f"Invalid arguments")

    @pytest.mark.parametrize(
        "vs_src, waf",
        [(waf_spec_vs_src, waf_pol_dataguard_src), (waf_route_vs_src, waf_pol_dataguard_src),],
    )
    def test_ap_waf_policy_allow(
        self,
        kube_apis,
        crd_ingress_controller_with_ap,
        virtual_server_setup,
        appprotect_setup,
        test_namespace,
        vs_src,
        waf,
    ):
        """
        Test waf policy when disabled
        """
        print(f"Create waf policy")
        create_ap_waf_policy_from_yaml(
            kube_apis.custom_objects,
            waf,
            test_namespace,
            test_namespace,
            False,
            False,
            ap_pol_name,
            log_name,
            "syslog:server=127.0.0.1:514",
        )
        wait_before_test()
        print(f"Patch vs with policy: {vs_src}")
        patch_virtual_server_from_yaml(
            kube_apis.custom_objects,
            virtual_server_setup.vs_name,
            vs_src,
            virtual_server_setup.namespace,
        )
        wait_before_test()
        ap_crd_info = read_ap_custom_resource(
            kube_apis.custom_objects, test_namespace, "appolicies", ap_policy_uds
        )
        assert_ap_crd_info(ap_crd_info, ap_policy_uds)
        wait_before_test(120)

        print(
            "----------------------- Send request with embedded malicious script----------------------"
        )
        response1 = requests.get(
            virtual_server_setup.backend_1_url + "</script>",
            headers={"host": virtual_server_setup.vs_host},
        )
        print(response1.text)

        print(
            "----------------------- Send request with blocked keyword in UDS----------------------"
        )
        response2 = requests.get(
            virtual_server_setup.backend_1_url,
            headers={"host": virtual_server_setup.vs_host},
            data="kic",
        )
        print(response2.text)

        delete_policy(kube_apis.custom_objects, "waf-policy", test_namespace)
        self.restore_default_vs(kube_apis, virtual_server_setup)
        assert_valid_responses(response1)
        assert_valid_responses(response2)

    @pytest.mark.flaky(max_runs=3)
    def test_ap_waf_policy_logs(
        self,
        kube_apis,
        crd_ingress_controller_with_ap,
        virtual_server_setup,
        appprotect_setup,
        test_namespace,
    ):
        """
        Test waf policy logs
        """
        src_syslog_yaml = f"{TEST_DATA}/ap-waf/syslog.yaml"
        log_loc = f"/var/log/messages"
        create_items_from_yaml(kube_apis, src_syslog_yaml, test_namespace)
        syslog_dst = f"syslog-svc.{test_namespace}"
        syslog_pod = kube_apis.v1.list_namespaced_pod(test_namespace).items[-1].metadata.name
        print(f"Create waf policy")
        create_ap_waf_policy_from_yaml(
            kube_apis.custom_objects,
            waf_pol_dataguard_src,
            test_namespace,
            test_namespace,
            True,
            True,
            ap_pol_name,
            log_name,
            f"syslog:server={syslog_dst}:514",
        )
        wait_before_test()
        print(f"Patch vs with policy: {waf_spec_vs_src}")
        patch_virtual_server_from_yaml(
            kube_apis.custom_objects,
            virtual_server_setup.vs_name,
            waf_spec_vs_src,
            virtual_server_setup.namespace,
        )
        wait_before_test()
        ap_crd_info = read_ap_custom_resource(
            kube_apis.custom_objects, test_namespace, "appolicies", ap_policy_uds
        )
        assert_ap_crd_info(ap_crd_info, ap_policy_uds)
        wait_before_test(120)

        print(
            "----------------------- Send request with embedded malicious script----------------------"
        )
        response = requests.get(
            virtual_server_setup.backend_1_url + "</script>",
            headers={"host": virtual_server_setup.vs_host},
        )
        print(response.text)
        log_contents = ""
        retry = 0
        while "ASM:attack_type" not in log_contents and retry <= 30:
            log_contents = get_file_contents(
                kube_apis.v1, log_loc, syslog_pod, test_namespace
            )
            retry += 1
            wait_before_test(1)
            print(f"Security log not updated, retrying... #{retry}")

        delete_policy(kube_apis.custom_objects, "waf-policy", test_namespace)
        self.restore_default_vs(kube_apis, virtual_server_setup)

        assert_invalid_responses(response)
        assert (
            f'ASM:attack_type="Non-browser Client,Abuse of Functionality,Cross Site Scripting (XSS)"'
            in log_contents
        )
        assert f'severity="Critical"' in log_contents
        assert f'request_status="blocked"' in log_contents
        assert f'outcome="REJECTED"' in log_contents


@pytest.mark.skip_for_nginx_oss
@pytest.mark.appprotect
@pytest.mark.parametrize(
    "crd_ingress_controller_with_ap, v_s_route_setup",
    [
        (
            {
                "type": "complete",
                "extra_args": [
                    f"-enable-custom-resources",
                    f"-enable-leader-election=false",
                    f"-enable-app-protect",
                ],
            },
            {"example": "virtual-server-route"},
        )
    ],
    indirect=True,
)
class TestAppProtectWAFPolicyVSR:
    def restore_default_vsr(self, kube_apis, v_s_route_setup) -> None:
        """
        Function to revert vsr deployments to standard state
        """
        patch_src_m = f"{TEST_DATA}/virtual-server-route/route-multiple.yaml"
        patch_v_s_route_from_yaml(
            kube_apis.custom_objects,
            v_s_route_setup.route_m.name,
            patch_src_m,
            v_s_route_setup.route_m.namespace,
        )
        wait_before_test()

    @pytest.mark.parametrize(
        "ap_enable",
        [
            True, 
            # False
        ],
    )
    def test_ap_waf_policy_block(
        self,
        kube_apis,
        crd_ingress_controller_with_ap,
        v_s_route_setup,
        appprotect_setup,
        test_namespace,
        ap_enable,
    ):
        """
        Test if WAF policy is working with VSR deployments
        """
        req_url = f"http://{v_s_route_setup.public_endpoint.public_ip}:{v_s_route_setup.public_endpoint.port}"

        print(f"Create waf policy")
        create_ap_waf_policy_from_yaml(
            kube_apis.custom_objects,
            waf_pol_dataguard_src,
            v_s_route_setup.route_m.namespace,
            test_namespace,
            ap_enable,
            ap_enable,
            ap_pol_name,
            log_name,
            "syslog:server=127.0.0.1:514",
        )
        wait_before_test()
        print(f"Patch vsr with policy: {waf_subroute_vsr_src}")
        patch_v_s_route_from_yaml(
            kube_apis.custom_objects,
            v_s_route_setup.route_m.name,
            waf_subroute_vsr_src,
            v_s_route_setup.route_m.namespace,
        )
        wait_before_test()
        ap_crd_info = read_ap_custom_resource(
            kube_apis.custom_objects, test_namespace, "appolicies", ap_policy_uds
        )
        assert_ap_crd_info(ap_crd_info, ap_policy_uds)
        wait_before_test(120)
        response = requests.get(
            f"{req_url}{v_s_route_setup.route_m.paths[0]}+'</script>'",
            headers={"host": v_s_route_setup.vs_host},
        )
        print(response.text)
        delete_policy(kube_apis.custom_objects, "waf-policy", v_s_route_setup.route_m.namespace)
        self.restore_default_vsr(kube_apis, v_s_route_setup)
        if ap_enable == True:
            assert_invalid_responses(response)
        elif ap_enable == False:
            assert_valid_responses(response)
        else:
            pytest.fail(f"Invalid arguments")
