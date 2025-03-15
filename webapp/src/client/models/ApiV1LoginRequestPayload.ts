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
/**
 * 
 * @export
 * @interface ApiV1LoginRequestPayload
 */
export interface ApiV1LoginRequestPayload {
    /**
     * 
     * @type {string}
     * @memberof ApiV1LoginRequestPayload
     */
    password?: string;
    /**
     * 
     * @type {boolean}
     * @memberof ApiV1LoginRequestPayload
     */
    rememberMe?: boolean;
    /**
     * 
     * @type {string}
     * @memberof ApiV1LoginRequestPayload
     */
    username?: string;
}

/**
 * Check if a given object implements the ApiV1LoginRequestPayload interface.
 */
export function instanceOfApiV1LoginRequestPayload(value: object): value is ApiV1LoginRequestPayload {
    return true;
}

export function ApiV1LoginRequestPayloadFromJSON(json: any): ApiV1LoginRequestPayload {
    return ApiV1LoginRequestPayloadFromJSONTyped(json, false);
}

export function ApiV1LoginRequestPayloadFromJSONTyped(json: any, ignoreDiscriminator: boolean): ApiV1LoginRequestPayload {
    if (json == null) {
        return json;
    }
    return {
        
        'password': json['password'] == null ? undefined : json['password'],
        'rememberMe': json['remember_me'] == null ? undefined : json['remember_me'],
        'username': json['username'] == null ? undefined : json['username'],
    };
}

export function ApiV1LoginRequestPayloadToJSON(json: any): ApiV1LoginRequestPayload {
    return ApiV1LoginRequestPayloadToJSONTyped(json, false);
}

export function ApiV1LoginRequestPayloadToJSONTyped(value?: ApiV1LoginRequestPayload | null, ignoreDiscriminator: boolean = false): any {
    if (value == null) {
        return value;
    }

    return {
        
        'password': value['password'],
        'remember_me': value['rememberMe'],
        'username': value['username'],
    };
}

