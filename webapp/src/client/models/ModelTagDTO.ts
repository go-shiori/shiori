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
 * @interface ModelTagDTO
 */
export interface ModelTagDTO {
    /**
     * Number of bookmarks with this tag
     * @type {number}
     * @memberof ModelTagDTO
     */
    bookmarkCount?: number;
    /**
     * Marks when a tag is deleted from a bookmark
     * @type {boolean}
     * @memberof ModelTagDTO
     */
    deleted?: boolean;
    /**
     * 
     * @type {number}
     * @memberof ModelTagDTO
     */
    id?: number;
    /**
     * 
     * @type {string}
     * @memberof ModelTagDTO
     */
    name?: string;
}

/**
 * Check if a given object implements the ModelTagDTO interface.
 */
export function instanceOfModelTagDTO(value: object): value is ModelTagDTO {
    return true;
}

export function ModelTagDTOFromJSON(json: any): ModelTagDTO {
    return ModelTagDTOFromJSONTyped(json, false);
}

export function ModelTagDTOFromJSONTyped(json: any, ignoreDiscriminator: boolean): ModelTagDTO {
    if (json == null) {
        return json;
    }
    return {
        
        'bookmarkCount': json['bookmark_count'] == null ? undefined : json['bookmark_count'],
        'deleted': json['deleted'] == null ? undefined : json['deleted'],
        'id': json['id'] == null ? undefined : json['id'],
        'name': json['name'] == null ? undefined : json['name'],
    };
}

export function ModelTagDTOToJSON(json: any): ModelTagDTO {
    return ModelTagDTOToJSONTyped(json, false);
}

export function ModelTagDTOToJSONTyped(value?: ModelTagDTO | null, ignoreDiscriminator: boolean = false): any {
    if (value == null) {
        return value;
    }

    return {
        
        'bookmark_count': value['bookmarkCount'],
        'deleted': value['deleted'],
        'id': value['id'],
        'name': value['name'],
    };
}

