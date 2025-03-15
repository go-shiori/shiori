/* tslint:disable */
/* eslint-disable */
/**
 * Shiori API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { mapValues } from '../runtime';
import type { ModelUserConfig } from './ModelUserConfig';
import {
    ModelUserConfigFromJSON,
    ModelUserConfigFromJSONTyped,
    ModelUserConfigToJSON,
    ModelUserConfigToJSONTyped,
} from './ModelUserConfig';

/**
 * 
 * @export
 * @interface ApiV1UpdateAccountPayload
 */
export interface ApiV1UpdateAccountPayload {
    /**
     * 
     * @type {ModelUserConfig}
     * @memberof ApiV1UpdateAccountPayload
     */
    config?: ModelUserConfig;
    /**
     * 
     * @type {string}
     * @memberof ApiV1UpdateAccountPayload
     */
    newPassword?: string;
    /**
     * 
     * @type {string}
     * @memberof ApiV1UpdateAccountPayload
     */
    oldPassword?: string;
    /**
     * 
     * @type {boolean}
     * @memberof ApiV1UpdateAccountPayload
     */
    owner?: boolean;
    /**
     * 
     * @type {string}
     * @memberof ApiV1UpdateAccountPayload
     */
    username?: string;
}

/**
 * Check if a given object implements the ApiV1UpdateAccountPayload interface.
 */
export function instanceOfApiV1UpdateAccountPayload(value: object): value is ApiV1UpdateAccountPayload {
    return true;
}

export function ApiV1UpdateAccountPayloadFromJSON(json: any): ApiV1UpdateAccountPayload {
    return ApiV1UpdateAccountPayloadFromJSONTyped(json, false);
}

export function ApiV1UpdateAccountPayloadFromJSONTyped(json: any, ignoreDiscriminator: boolean): ApiV1UpdateAccountPayload {
    if (json == null) {
        return json;
    }
    return {
        
        'config': json['config'] == null ? undefined : ModelUserConfigFromJSON(json['config']),
        'newPassword': json['new_password'] == null ? undefined : json['new_password'],
        'oldPassword': json['old_password'] == null ? undefined : json['old_password'],
        'owner': json['owner'] == null ? undefined : json['owner'],
        'username': json['username'] == null ? undefined : json['username'],
    };
}

export function ApiV1UpdateAccountPayloadToJSON(json: any): ApiV1UpdateAccountPayload {
    return ApiV1UpdateAccountPayloadToJSONTyped(json, false);
}

export function ApiV1UpdateAccountPayloadToJSONTyped(value?: ApiV1UpdateAccountPayload | null, ignoreDiscriminator: boolean = false): any {
    if (value == null) {
        return value;
    }

    return {
        
        'config': ModelUserConfigToJSON(value['config']),
        'new_password': value['newPassword'],
        'old_password': value['oldPassword'],
        'owner': value['owner'],
        'username': value['username'],
    };
}

