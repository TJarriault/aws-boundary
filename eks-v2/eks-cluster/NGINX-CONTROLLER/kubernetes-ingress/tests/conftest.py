"""Describe overall framework configuration."""

import os

import pytest
from kubernetes.config.kube_config import KUBE_CONFIG_DEFAULT_LOCATION
from settings import (BATCH_RESOURCES, BATCH_START, DEFAULT_DEPLOYMENT_TYPE,
                      DEFAULT_IC_TYPE, DEFAULT_IMAGE, DEFAULT_PULL_POLICY,
                      DEFAULT_SERVICE, NUM_REPLICAS)
from suite.resources_utils import get_first_pod_name


def pytest_addoption(parser) -> None:
    """Get cli-arguments.

    :param parser: pytest parser
    :return:
    """
    parser.addoption(
        "--context",
        action="store",
        default="",
        help="The context to use in the kubeconfig file.",
    )
    parser.addoption(
        "--image",
        action="store",
        default=DEFAULT_IMAGE,
        help="The Ingress Controller image.",
    )
    parser.addoption(
        "--image-pull-policy",
        action="store",
        default=DEFAULT_PULL_POLICY,
        help="The pull policy of the Ingress Controller image.",
    )
    parser.addoption(
        "--deployment-type",
        action="store",
        default=DEFAULT_DEPLOYMENT_TYPE,
        help="The type of the IC deployment: deployment or daemon-set.",
    )
    parser.addoption(
        "--ic-type",
        action="store",
        default=DEFAULT_IC_TYPE,
        help="The type of the Ingress Controller: nginx-ingress or nginx-plus-ingress.",
    )
    parser.addoption(
        "--service",
        action="store",
        default=DEFAULT_SERVICE,
        help="The type of the Ingress Controller service: nodeport or loadbalancer.",
    )
    parser.addoption(
        "--replicas",
        action="store",
        default=NUM_REPLICAS,
        help="Number of replica pods for type deployment",
    )
    parser.addoption(
        "--node-ip",
        action="store",
        help="The public IP of a cluster node. Not required if you use the loadbalancer service (see --service argument).",
    )
    parser.addoption(
        "--kubeconfig",
        action="store",
        default=os.path.expanduser(KUBE_CONFIG_DEFAULT_LOCATION),
        help="An absolute path to a kubeconfig file.",
    )
    parser.addoption(
        "--show-ic-logs",
        action="store",
        default="no",
        help="Show IC logs in stdout on test failure",
    )
    parser.addoption(
        "--batch-start",
        action="store",
        default=BATCH_START,
        help="Run tests for pods restarts with multiple resources deployed (Ingress/VS): True/False",
    )
    parser.addoption(
        "--batch-resources",
        action="store",
        default=BATCH_RESOURCES,
        help="Number of VS/Ingress resources to deploy",
    )


# import fixtures into pytest global namespace
pytest_plugins = ["suite.fixtures"]


def pytest_collection_modifyitems(config, items) -> None:
    """
    Skip tests marked with '@pytest.mark.skip_for_nginx_oss' for Nginx OSS runs.
    Skip tests marked with '@pytest.mark.appprotect' for non AP images.

    :param config: pytest config
    :param items: pytest collected test-items
    :return:
    """
    if config.getoption("--ic-type") == "nginx-ingress":
        skip_for_nginx_oss = pytest.mark.skip(reason="Skip a test for Nginx OSS")
        for item in items:
            if "skip_for_nginx_oss" in item.keywords:
                item.add_marker(skip_for_nginx_oss)
    if config.getoption("--ic-type") == "nginx-plus-ingress":
        skip_for_nginx_plus = pytest.mark.skip(reason="Skip a test for Nginx Plus")
        for item in items:
            if "skip_for_nginx_plus" in item.keywords:
                item.add_marker(skip_for_nginx_plus)
    if config.getoption("--service") == "loadbalancer":
        skip_for_loadbalancer = pytest.mark.skip(reason="Skip a test for loadbalancer service")
        for item in items:
            if "skip_for_loadbalancer" in item.keywords:
                item.add_marker(skip_for_loadbalancer)
    if "-ap" not in config.getoption("--image"):
        appprotect = pytest.mark.skip(reason="Skip AppProtect test in non-AP image")
        for item in items:
            if "appprotect" in item.keywords:
                item.add_marker(appprotect)
    if "-dos" not in config.getoption("--image"):
        dos = pytest.mark.skip(reason="Skip DOS test in non-DOS image")
        for item in items:
            if "dos" in item.keywords:
                item.add_marker(dos)
    if  str(config.getoption("--batch-start")) != "True":
        batch_start = pytest.mark.skip(reason="Skipping pod restart test with multiple resources")
        for item in items:
            if "batch_start" in item.keywords:
                item.add_marker(batch_start)


@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item) -> None:
    """
    Print out IC Pod logs on test failure.

    Only look at actual failing test calls, not setup/teardown.
    Only show the logs if commandline argument `--show-ic-logs` is set to 'yes'

    :param item:
    :return:
    """
    # execute all other hooks to obtain the report object
    outcome = yield
    rep = outcome.get_result()

    # we only look at actual failing test calls, not setup/teardown
    if (
        rep.when == "call"
        and rep.failed
        and item.config.getoption("--show-ic-logs") == "yes"
    ):
        pod_namespace = item.funcargs["ingress_controller_prerequisites"].namespace
        pod_name = get_first_pod_name(item.funcargs["kube_apis"].v1, pod_namespace)
        print("\n===================== IC Logs Start =====================")
        print(
            item.funcargs["kube_apis"].v1.read_namespaced_pod_log(
                pod_name, pod_namespace
            )
        )
        print("\n===================== IC Logs End =====================")
