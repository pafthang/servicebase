import { describe, assert, test, beforeAll, afterAll, afterEach } from "vitest";
import { FetchMock } from "../mocks";
import { crudServiceTestsSuite } from "../suites";
import Client from "@/Client";
import { TeamService } from "@/services/TeamService";

describe("TeamService", function () {
    const client = new Client("test_base_url");
    const service = new TeamService(client);

    crudServiceTestsSuite(service, "/api/teams");

    const fetchMock = new FetchMock();

    beforeAll(function () {
        fetchMock.init();
    });

    afterAll(function () {
        fetchMock.restore();
    });

    afterEach(function () {
        fetchMock.clearMocks();
    });

    describe("getMyTeams()", function () {
        test("Should fetch auth user teams", async function () {
            const replyBody = [{ id: "t1", name: "admin" }];

            fetchMock.on({
                method: "GET",
                url: service.client.buildUrl("/api/teams/me") + "?q1=123",
                additionalMatcher: (_, config) => {
                    return config?.headers?.["x-test"] === "123";
                },
                replyCode: 200,
                replyBody: replyBody,
            });

            const result = await service.getMyTeams({
                q1: 123,
                headers: { "x-test": "123" },
            });

            assert.deepEqual(result, replyBody);
        });
    });

    describe("getMembers()", function () {
        test("Should fetch all team members", async function () {
            const replyBody = [
                {
                    membership: { id: "m1", team: "t1", userId: "u1", userCollectionId: "users" },
                    user: { id: "u1", collectionId: "users", collectionName: "users" },
                },
            ];

            fetchMock.on({
                method: "GET",
                url: service.client.buildUrl("/api/teams/%40team/members") + "?q1=123",
                additionalMatcher: (_, config) => {
                    return config?.headers?.["x-test"] === "123";
                },
                replyCode: 200,
                replyBody: replyBody,
            });

            const result = await service.getMembers("@team", {
                q1: 123,
                headers: { "x-test": "123" },
            });

            assert.deepEqual(result, replyBody);
        });
    });

    describe("addMember()", function () {
        test("Should add member to a team", async function () {
            const replyBody = {
                membership: { id: "m1", team: "t1", userId: "u1", userCollectionId: "users" },
                user: { id: "u1", collectionId: "users", collectionName: "users" },
            };

            fetchMock.on({
                method: "POST",
                url: service.client.buildUrl("/api/teams/%40team/members") + "?q1=123",
                body: { userId: "@user" },
                additionalMatcher: (_, config) => {
                    return config?.headers?.["x-test"] === "123";
                },
                replyCode: 200,
                replyBody: replyBody,
            });

            const result = await service.addMember("@team", "@user", {
                q1: 123,
                headers: { "x-test": "123" },
            });

            assert.deepEqual(result, replyBody);
        });
    });

    describe("removeMember()", function () {
        test("Should remove member from a team", async function () {
            fetchMock.on({
                method: "DELETE",
                url: service.client.buildUrl("/api/teams/%40team/members/%40user") + "?q1=123",
                additionalMatcher: (_, config) => {
                    return config?.headers?.["x-test"] === "123";
                },
                replyCode: 204,
                replyBody: true,
            });

            const result = await service.removeMember("@team", "@user", {
                q1: 123,
                headers: { "x-test": "123" },
            });

            assert.deepEqual(result, true);
        });
    });
});
