import requests
import pytest
import subprocess
import os

from settings import TEST_DATA
from suite.custom_resources_utils import (
    create_dos_logconf_from_yaml,
    create_dos_policy_from_yaml,
    create_dos_protected_from_yaml,
    delete_dos_policy,
    delete_dos_logconf,
    delete_dos_protected,
)
from suite.dos_utils import find_in_log, log_content_to_dic
from suite.resources_utils import (
    wait_before_test,
    create_example_app,
    wait_until_all_pods_are_ready,
    create_items_from_yaml,
    delete_items_from_yaml,
    delete_common_app,
    ensure_connection_to_public_endpoint,
    create_ingress_with_dos_annotations,
    ensure_response_from_backend,
    get_ingress_nginx_template_conf,
    get_file_contents,
    get_test_file_name,
    write_to_json,
    replace_configmap_from_yaml,
    scale_deployment,
    nginx_reload,
    get_pods_amount,
    get_pods_amount_with_name,
    clear_file_contents,
)
from suite.yaml_utils import get_first_ingress_host_from_yaml
from datetime import datetime

src_ing_yaml = f"{TEST_DATA}/dos/dos-ingress.yaml"
valid_resp_addr = "Server address:"
valid_resp_name = "Server name:"
invalid_resp_title = "Request Rejected"
invalid_resp_body = "The requested URL was rejected. Please consult with your administrator."
reload_times = {}

class DosSetup:
    """
    Encapsulate the example details.
    Attributes:
        req_url (str):
        pol_name (str):
        log_name (str):
    """
    def __init__(self, req_url, pol_name, log_name):
        self.req_url = req_url
        self.pol_name = pol_name
        self.log_name = log_name


@pytest.fixture(scope="class")
def dos_setup(
        request, kube_apis, ingress_controller_endpoint, ingress_controller_prerequisites, test_namespace
) -> DosSetup:
    """
    Deploy simple application and all the DOS resources under test in one namespace.

    :param request: pytest fixture
    :param kube_apis: client apis
    :param ingress_controller_endpoint: public endpoint
    :param ingress_controller_prerequisites: IC pre-requisites
    :param test_namespace:
    :return: DosSetup
    """

    print(f"------------- Replace ConfigMap --------------")
    replace_configmap_from_yaml(
        kube_apis.v1,
        ingress_controller_prerequisites.config_map["metadata"]["name"],
        ingress_controller_prerequisites.namespace,
        f"{TEST_DATA}/dos/nginx-config.yaml"
    )

    print("------------------------- Deploy Dos backend application -------------------------")
    create_example_app(kube_apis, "dos", test_namespace)
    req_url = f"http://{ingress_controller_endpoint.public_ip}:{ingress_controller_endpoint.port}/"
    wait_until_all_pods_are_ready(kube_apis.v1, test_namespace)
    ensure_connection_to_public_endpoint(
        ingress_controller_endpoint.public_ip,
        ingress_controller_endpoint.port,
        ingress_controller_endpoint.port_ssl,
    )

    print("------------------------- Deploy Secret -----------------------------")
    src_sec_yaml = f"{TEST_DATA}/dos/dos-secret.yaml"
    create_items_from_yaml(kube_apis, src_sec_yaml, test_namespace)

    print("------------------------- Deploy logconf -----------------------------")
    src_log_yaml = f"{TEST_DATA}/dos/dos-logconf.yaml"
    log_name = create_dos_logconf_from_yaml(kube_apis.custom_objects, src_log_yaml, test_namespace)

    print(f"------------------------- Deploy dospolicy ---------------------------")
    src_pol_yaml = f"{TEST_DATA}/dos/dos-policy.yaml"
    pol_name = create_dos_policy_from_yaml(kube_apis.custom_objects, src_pol_yaml, test_namespace)

    print(f"------------------------- Deploy protected resource ---------------------------")
    src_protected_yaml = f"{TEST_DATA}/dos/dos-protected.yaml"
    protected_name = create_dos_protected_from_yaml(kube_apis.custom_objects, src_protected_yaml, test_namespace, ingress_controller_prerequisites.namespace)

    for item in kube_apis.v1.list_namespaced_pod(ingress_controller_prerequisites.namespace).items:
        if "nginx-ingress" in item.metadata.name:
            nginx_reload(kube_apis.v1, item.metadata.name, ingress_controller_prerequisites.namespace)

    def fin():
        print("Clean up:")
        delete_dos_policy(kube_apis.custom_objects, pol_name, test_namespace)
        delete_dos_logconf(kube_apis.custom_objects, log_name, test_namespace)
        delete_dos_protected(kube_apis.custom_objects, protected_name, test_namespace)
        delete_common_app(kube_apis, "dos", test_namespace)
        delete_items_from_yaml(kube_apis, src_sec_yaml, test_namespace)
        write_to_json(f"reload-{get_test_file_name(request.node.fspath)}.json", reload_times)

    request.addfinalizer(fin)

    return DosSetup(req_url, pol_name, log_name)


@pytest.mark.dos
@pytest.mark.parametrize(
    "crd_ingress_controller_with_dos",
    [
        {
            "extra_args": [
                f"-enable-custom-resources",
                f"-enable-app-protect-dos",
                f"-v=3",
            ]
        }
    ],
    indirect=["crd_ingress_controller_with_dos"],
)
class TestDos:
    def getPodNameThatContains(self, kube_apis, namespace, contains_string):
        for item in kube_apis.v1.list_namespaced_pod(namespace).items:
            if contains_string in item.metadata.name:
                return item.metadata.name
        return ""

    def test_ap_nginx_config_entries(
            self, kube_apis, ingress_controller_prerequisites, crd_ingress_controller_with_dos, dos_setup, test_namespace
    ):
        """
        Test to verify Dos directive in nginx config
        """
        conf_directive = [
            f"app_protect_dos_enable on;",
            f"app_protect_dos_security_log_enable on;",
            f"app_protect_dos_monitor uri=dos.example.com protocol=http1 timeout=5;",
            f"app_protect_dos_name \"{test_namespace}/dos-protected/name\";",
            f"app_protect_dos_policy_file /etc/nginx/dos/policies/{test_namespace}_{dos_setup.pol_name}.json;",
            f"app_protect_dos_security_log_enable on;",
            f"app_protect_dos_security_log /etc/nginx/dos/logconfs/{test_namespace}_{dos_setup.log_name}.json syslog:server=syslog-svc.{ingress_controller_prerequisites.namespace}.svc.cluster.local:514;",
        ]

        create_ingress_with_dos_annotations(
            kube_apis, src_ing_yaml, test_namespace, test_namespace + "/dos-protected",
        )

        ingress_host = get_first_ingress_host_from_yaml(src_ing_yaml)
        ensure_response_from_backend(dos_setup.req_url, ingress_host, check404=True)

        pod_name = self.getPodNameThatContains(kube_apis, ingress_controller_prerequisites.namespace, "nginx-ingress")

        result_conf = get_ingress_nginx_template_conf(
            kube_apis.v1, test_namespace, "dos-ingress", pod_name, "nginx-ingress"
        )

        delete_items_from_yaml(kube_apis, src_ing_yaml, test_namespace)

        for _ in conf_directive:
            assert _ in result_conf

    def test_dos_sec_logs_on(
            self,
            kube_apis,
            ingress_controller_prerequisites,
            crd_ingress_controller_with_dos,
            dos_setup,
            test_namespace,
    ):
        """
        Test corresponding log entries with correct policy (includes setting up a syslog server as defined in syslog.yaml)
        """
        print("----------------------- Get syslog pod name ----------------------")
        syslog_pod = self.getPodNameThatContains(kube_apis, ingress_controller_prerequisites.namespace, "syslog")
        assert "syslog" in syslog_pod

        log_loc = f"/var/log/messages"
        clear_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)

        create_ingress_with_dos_annotations(
            kube_apis, src_ing_yaml, test_namespace, test_namespace+"/dos-protected"
        )
        ingress_host = get_first_ingress_host_from_yaml(src_ing_yaml)

        print("--------- Run test while DOS module is enabled with correct policy ---------")

        ensure_response_from_backend(dos_setup.req_url, ingress_host, check404=True)
        pod_name = self.getPodNameThatContains(kube_apis, ingress_controller_prerequisites.namespace, "nginx-ingress")

        get_ingress_nginx_template_conf(
            kube_apis.v1, test_namespace, "dos-ingress", pod_name, "nginx-ingress"
        )

        print("----------------------- Send request ----------------------")
        response = requests.get(
            dos_setup.req_url, headers={"host": "dos.example.com"}, verify=False
        )
        print(response.text)
        wait_before_test(10)

        print(f'log_loc {log_loc} syslog_pod {syslog_pod} namespace {ingress_controller_prerequisites.namespace}')
        log_contents = get_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)

        delete_items_from_yaml(kube_apis, src_ing_yaml, test_namespace)

        print(log_contents)

        assert 'product="app-protect-dos"' in log_contents
        assert f'vs_name="{test_namespace}/dos-protected/name"' in log_contents
        assert 'bad_actor' in log_contents

    @pytest.mark.skip
    def test_dos_under_attack_no_learning(
            self, kube_apis, ingress_controller_prerequisites, crd_ingress_controller_with_dos, dos_setup, test_namespace
    ):
        """
        Test App Protect Dos: Block bad clients attack
        """
        log_loc = f"/var/log/messages"
        print("----------------------- Get syslog pod name ----------------------")
        syslog_pod = self.getPodNameThatContains(kube_apis, ingress_controller_prerequisites.namespace, "syslog")
        assert "syslog" in syslog_pod
        clear_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)

        print("------------------------- Deploy ingress -----------------------------")
        create_ingress_with_dos_annotations(
            kube_apis, src_ing_yaml, test_namespace, test_namespace+"/dos-protected"
        )
        ingress_host = get_first_ingress_host_from_yaml(src_ing_yaml)

        print("------------------------- Attack -----------------------------")
        wait_before_test(10)
        print("start bad clients requests")
        p_attack = subprocess.Popen(
            [f"exec {TEST_DATA}/dos/bad_clients_xff.sh {ingress_host} {dos_setup.req_url}"],
            shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

        print("Attack for 30 seconds")
        wait_before_test(30)

        print("Stop Attack")
        p_attack.terminate()

        print("wait max 140 seconds after attack stop, to get attack ended")
        find_in_log(kube_apis, log_loc, syslog_pod, ingress_controller_prerequisites.namespace, 140, "attack_event=\"Attack ended\"")

        log_contents = get_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)
        log_info_dic = log_content_to_dic(log_contents)

        # Analyze the log
        no_attack = False
        attack_started = False
        under_attack = False
        attack_ended = False
        for log in log_info_dic:
            # Start with no attack
            if log['attack_event'] == "No Attack" and int(log['dos_attack_id']) == 0 and not no_attack:
                no_attack = True
            # Attack started
            elif log['attack_event'] == "Attack started" and int(log['dos_attack_id']) > 0 and not attack_started:
                attack_started = True
            # Under attack
            elif log['attack_event'] == "Under Attack" and int(log['dos_attack_id']) > 0 and not under_attack:
                under_attack = True
            # Attack ended
            elif log['attack_event'] == "Attack ended" and int(log['dos_attack_id']) > 0 and not attack_ended:
                attack_ended = True

        delete_items_from_yaml(kube_apis, src_ing_yaml, test_namespace)

        assert (
                no_attack
                and attack_started
                and under_attack
                and attack_ended
        )

    @pytest.mark.skip
    def test_dos_under_attack_with_learning(
            self, kube_apis, ingress_controller_prerequisites, crd_ingress_controller_with_dos, dos_setup, test_namespace
    ):
        """
        Test App Protect Dos: Block bad clients attack with learning
        """
        log_loc = f"/var/log/messages"
        print("----------------------- Get syslog pod name ----------------------")
        syslog_pod = self.getPodNameThatContains(kube_apis, ingress_controller_prerequisites.namespace, "syslog")
        assert "syslog" in syslog_pod
        clear_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)

        print("------------------------- Deploy ingress -----------------------------")
        create_ingress_with_dos_annotations(
            kube_apis, src_ing_yaml, test_namespace, test_namespace+"/dos-protected"
        )
        ingress_host = get_first_ingress_host_from_yaml(src_ing_yaml)

        print("------------------------- Learning Phase -----------------------------")
        print("start good clients requests")
        p_good_client = subprocess.Popen(
            [f"exec {TEST_DATA}/dos/good_clients_xff.sh {ingress_host} {dos_setup.req_url}"],
            preexec_fn=os.setsid, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

        print("Learning for max 10 minutes")
        find_in_log(kube_apis, log_loc, syslog_pod, ingress_controller_prerequisites.namespace, 600, "learning_confidence=\"Ready\"")

        print("------------------------- Attack -----------------------------")
        print("start bad clients requests")
        p_attack = subprocess.Popen(
            [f"exec {TEST_DATA}/dos/bad_clients_xff.sh {ingress_host} {dos_setup.req_url}"],
            preexec_fn=os.setsid, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

        print("Attack for 300 seconds")
        wait_before_test(300)

        print("Stop Attack")
        p_attack.terminate()

        print("wait max 140 seconds after attack stop, to get attack ended")
        find_in_log(kube_apis, log_loc, syslog_pod, ingress_controller_prerequisites.namespace, 140, "attack_event=\"Attack ended\"")

        print("Stop Good Client")
        p_good_client.terminate()

        log_contents = get_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)
        log_info_dic = log_content_to_dic(log_contents)

        # Analyze the log
        no_attack = False
        attack_started = False
        under_attack = False
        attack_ended = False
        bad_actor_detected = False
        signature_detected = False
        health_ok = False
        bad_ip = ['1.1.1.1', '1.1.1.2', '1.1.1.3']
        fmt = '%b %d %Y %H:%M:%S'
        for log in log_info_dic:
            if log['attack_event'] == 'No Attack':
                if int(log['dos_attack_id']) == 0 and not no_attack:
                    no_attack = True
            elif log['attack_event'] == 'Attack started':
                if int(log['dos_attack_id']) > 0 and not attack_started:
                    attack_started = True
                    start_attack_time = datetime.strptime(log['date_time'], fmt)
            elif log['attack_event'] == 'Under Attack':
                under_attack = True
                if not health_ok and float(log['stress_level']) < 0.6:
                    health_ok = True
                    health_ok_time = datetime.strptime(log['date_time'], fmt)
            elif log['attack_event'] == 'Attack signature detected':
                signature_detected = True
            elif log['attack_event'] == 'Bad actors detected':
                if under_attack:
                    bad_actor_detected = True
            elif log['attack_event'] == 'Bad actor detection':
                if under_attack and log['source_ip'] in bad_ip:
                    bad_ip.remove(log['source_ip'])
            elif log['attack_event'] == 'Attack ended':
                attack_ended = True

        delete_items_from_yaml(kube_apis, src_ing_yaml, test_namespace)

        assert (
                no_attack
                and attack_started
                and under_attack
                and attack_ended
                and health_ok
                and (health_ok_time - start_attack_time).total_seconds() < 150
                and signature_detected
                and bad_actor_detected
                and len(bad_ip) == 0
        )

    @pytest.mark.skip
    def test_dos_arbitrator(
            self, kube_apis, ingress_controller_prerequisites, crd_ingress_controller_with_dos, dos_setup,
            test_namespace
    ):
        """
        Test App Protect Dos: Check new IC pod get learning info
        """
        print("----------------------- Get syslog pod name ----------------------")
        syslog_pod = self.getPodNameThatContains(kube_apis, ingress_controller_prerequisites.namespace, "syslog")
        assert "syslog" in syslog_pod
        log_loc = f"/var/log/messages"
        clear_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)

        print("------------------------- Deploy ingress -----------------------------")
        create_ingress_with_dos_annotations(
            kube_apis, src_ing_yaml, test_namespace, test_namespace+"/dos-protected"
        )
        ingress_host = get_first_ingress_host_from_yaml(src_ing_yaml)

        # print("------------------------- Learning Phase -----------------------------")
        print("start good clients requests")
        p_good_client = subprocess.Popen(
            [f"exec {TEST_DATA}/dos/good_clients_xff.sh {ingress_host} {dos_setup.req_url}"],
            preexec_fn=os.setsid, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

        print("Learning for max 10 minutes")
        find_in_log(kube_apis, log_loc, syslog_pod, ingress_controller_prerequisites.namespace, 600, "learning_confidence=\"Ready\"")

        print("------------------------- Check new IC pod get info from arbitrator -----------------------------")
        ic_ns = ingress_controller_prerequisites.namespace
        scale_deployment(kube_apis.v1, kube_apis.apps_v1_api, "nginx-ingress", ic_ns, 2)
        while get_pods_amount_with_name(kube_apis.v1, "nginx-ingress", "nginx-ingress") is not 2:
            print(f"Number of replicas is not 2, retrying...")
            wait_before_test()

        print("------------------------- Check if new pod receive info from arbitrator -----------------------------")
        print("Wait for 30 seconds")
        wait_before_test(30)

        log_contents = get_file_contents(kube_apis.v1, log_loc, syslog_pod, ingress_controller_prerequisites.namespace)
        log_info_dic = log_content_to_dic(log_contents)

        print("Stop Good Client")
        p_good_client.terminate()

        learning_units_hostname = []
        for log in log_info_dic:
            if log['unit_hostname'] not in learning_units_hostname and log['learning_confidence'] == "Ready":
                learning_units_hostname.append(log['unit_hostname'])

        delete_items_from_yaml(kube_apis, src_ing_yaml, test_namespace)

        assert (
                len(learning_units_hostname) == 2
        )
