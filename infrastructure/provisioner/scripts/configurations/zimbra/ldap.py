import sys
import os

# Get the parent directory
parent_dir = os.getcwd()
# # Add the parent directory to sys.path
sys.path.append(parent_dir)

from lib.utilities.auto_configuration import (
    get_resources_state,
    post_provisioning_configuration,
    freeipa_resources_state,
    get_command_args,
    log,
    concatenate_domain,
    remove_first_segment,
)


def main(args):
    # zimbra
    zimbra_reference = args.reference
    zimbra_state = get_resources_state(zimbra_reference)["data"]
    zimbra_domain = zimbra_state["State"]["ovh_domain_zone_record"]["values"]
    zimbra_domain = concatenate_domain(
        sub_domain=zimbra_domain["subdomain"],
        root_domain=zimbra_domain["zone"],
    )

    # zimbra MX
    zimbra_post_body = zimbra_state["Job"]["PostBody"]
    zimbra_metadata = zimbra_post_body["platform"]["metadata"]
    zimbra_mx_domain = zimbra_metadata.get("mx_domain")

    if not zimbra_mx_domain:
        if zimbra_post_body.get("mx_domain_value"):
            zimbra_mx_domain_value = zimbra_post_body["mx_domain_value"]
            zimbra_mx_domain = concatenate_domain(
                sub_domain=zimbra_mx_domain_value["subdomain"],
                root_domain=zimbra_mx_domain_value["zone"],
            )
        else:
            zimbra_mx_domain = remove_first_segment(zimbra_domain)

    # FreeIPA
    ipa_domain_dc, freeipa_credentials, freeipa_ipv4_address = freeipa_resources_state(
        args.config_reference
    )

    body = {
        "ref": zimbra_reference,
        "platform": {
            "name": args.platform,
            "metadata": {
                "zimbra_fqdn": zimbra_domain,
                "zimbra_domain": zimbra_mx_domain,
                "configuration_type": args.type,
                "configuration": {
                    "ldap_filter_username": "uid=%u",
                    "ldap_server_url": "ldap://%s:389" % freeipa_ipv4_address,
                    "ldap_search_base": "cn=accounts,%s" % ipa_domain_dc,
                    "ldap_bind_dn": "uid=admin,cn=users,cn=accounts,%s" % ipa_domain_dc,
                    "ldap_bind_password": freeipa_credentials["password"],
                },
            },
        },
    }

    response = post_provisioning_configuration(body)
    log(response["job"]["ID"])


if __name__ == "__main__":
    args = get_command_args()
    main(args)
